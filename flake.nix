{
  description = "UniFi Threat Sync - Automatically sync threat intelligence feeds to UniFi Network Application";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        
        # Package version - update this with your releases
        version = "1.0.0";
        
        # Build the Go application
        unifi-threat-sync = pkgs.buildGoModule {
          pname = "unifi-threat-sync";
          inherit version;
          
          src = ./.;
          
          # Vendoring Go dependencies
          # To update: nix run nixpkgs#go-mod-vendor -- .
          vendorHash = "sha256-VTZPgHFKNVxa6LGT8ykDjlsLDMhGi+FHWphNmUbsLG0=";
          
          # Build flags
          ldflags = [
            "-s"
            "-w"
            "-X main.version=${version}"
            "-X main.commit=${self.rev or "dev"}"
            "-X main.date=1970-01-01T00:00:00Z"
          ];
          
          # Disable CGO for static binary
          CGO_ENABLED = 0;
          
          # Build tags for production
          tags = [ "netgo" "osusergo" ];
          
          meta = with pkgs.lib; {
            description = "Automatically sync threat intelligence feeds to UniFi Network Application";
            homepage = "https://github.com/0x4272616E646F6E/unifi-threat-sync";
            license = licenses.mit;
            maintainers = [ ];
            mainProgram = "unifi-threat-sync";
          };
        };

        # Docker image
        dockerImage = pkgs.dockerTools.buildLayeredImage {
          name = "unifi-threat-sync";
          tag = version;
          
          contents = [ unifi-threat-sync ];
          
          config = {
            Cmd = [ "/bin/unifi-threat-sync" ];
            ExposedPorts = {
              "8080/tcp" = {};
            };
            Env = [
              "SSL_CERT_FILE=${pkgs.cacert}/etc/ssl/certs/ca-bundle.crt"
            ];
            Labels = {
              "org.opencontainers.image.source" = "https://github.com/0x4272616E646F6E/unifi-threat-sync";
              "org.opencontainers.image.description" = "UniFi Threat Sync";
              "org.opencontainers.image.licenses" = "MIT";
            };
            User = "65532:65532";
            WorkingDir = "/";
          };
        };

      in
      {
        # Default package
        packages.default = unifi-threat-sync;
        packages.unifi-threat-sync = unifi-threat-sync;
        packages.docker = dockerImage;

        # Development shell
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            # Go toolchain
            go_1_25
            gotools
            go-tools
            golangci-lint
            
            # Testing & coverage
            gotest
            gotestsum
            
            # Code generation & formatting
            gofumpt
            golines
            
            # Documentation
            godoc
            
            # Container tools
            docker
            docker-compose
            kubectl
            
            # Utilities
            gnumake
            git
            wget
            curl
            jq
          ];
          
          shellHook = ''
            echo "🚀 UniFi Threat Sync Development Environment"
            echo ""
            echo "Available commands:"
            echo "  make build          - Build the binary"
            echo "  make test           - Run tests"
            echo "  make lint           - Run linters"
            echo "  make docker-build   - Build Docker image"
            echo "  make help           - Show all make targets"
            echo ""
            echo "Go version: $(go version)"
            echo "Nix channel: nixos-25.05 (stable)"
            echo ""
          '';
          
          # Environment variables
          GOROOT = "${pkgs.go_1_25}/share/go";
          CGO_ENABLED = "0";
        };

        # Apps for easy running
        apps.default = flake-utils.lib.mkApp {
          drv = unifi-threat-sync;
        };

        apps.unifi-threat-sync = flake-utils.lib.mkApp {
          drv = unifi-threat-sync;
        };

        # Formatter
        formatter = pkgs.nixpkgs-fmt;
      }
    ) // {
      # NixOS module
      nixosModules.default = { config, lib, pkgs, ... }:
        with lib;
        let
          cfg = config.services.unifi-threat-sync;
          configFile = pkgs.writeText "unifi-threat-sync-config.yaml" (builtins.toJSON cfg.settings);
        in
        {
          options.services.unifi-threat-sync = {
            enable = mkEnableOption "UniFi Threat Sync service";

            package = mkOption {
              type = types.package;
              default = self.packages.${pkgs.system}.default;
              description = "The unifi-threat-sync package to use";
            };

            settings = mkOption {
              type = types.attrs;
              default = {};
              description = ''
                Configuration for UniFi Threat Sync.
                See https://github.com/0x4272616E646F6E/unifi-threat-sync for options.
              '';
              example = literalExpression ''
                {
                  unifi = {
                    url = "https://udm-pro.local";
                    username = "admin";
                    password = "secret";
                    site = "default";
                    groupName = "uts-block-list";
                  };
                  sync = {
                    interval = "60m";
                  };
                  health = {
                    enabled = true;
                    port = 8080;
                  };
                  feeds = [
                    {
                      name = "Spamhaus DROP";
                      url = "https://www.spamhaus.org/drop/drop.txt";
                      parser = "plain";
                      enabled = true;
                    }
                  ];
                }
              '';
            };

            environmentFile = mkOption {
              type = types.nullOr types.path;
              default = null;
              description = ''
                Environment file containing secrets (UNIFI_USER, UNIFI_PASS, etc.)
              '';
            };

            user = mkOption {
              type = types.str;
              default = "unifi-threat-sync";
              description = "User account under which unifi-threat-sync runs";
            };

            group = mkOption {
              type = types.str;
              default = "unifi-threat-sync";
              description = "Group under which unifi-threat-sync runs";
            };
          };

          config = mkIf cfg.enable {
            systemd.services.unifi-threat-sync = {
              description = "UniFi Threat Sync";
              after = [ "network-online.target" ];
              wants = [ "network-online.target" ];
              wantedBy = [ "multi-user.target" ];

              serviceConfig = {
                Type = "simple";
                User = cfg.user;
                Group = cfg.group;
                ExecStart = "${cfg.package}/bin/unifi-threat-sync -config ${configFile}";
                Restart = "on-failure";
                RestartSec = "10s";
                
                # Security hardening
                NoNewPrivileges = true;
                PrivateTmp = true;
                ProtectSystem = "strict";
                ProtectHome = true;
                ReadOnlyPaths = "/";
                ProtectKernelTunables = true;
                ProtectKernelModules = true;
                ProtectControlGroups = true;
                RestrictAddressFamilies = [ "AF_INET" "AF_INET6" ];
                RestrictNamespaces = true;
                LockPersonality = true;
                MemoryDenyWriteExecute = true;
                RestrictRealtime = true;
                RestrictSUIDSGID = true;
                PrivateMounts = true;
                
                # Load environment file if specified
                EnvironmentFile = mkIf (cfg.environmentFile != null) cfg.environmentFile;
              };
            };

            users.users.${cfg.user} = {
              isSystemUser = true;
              group = cfg.group;
              description = "UniFi Threat Sync service user";
            };

            users.groups.${cfg.group} = {};
          };
        };

      # Home Manager module
      homeManagerModules.default = { config, lib, pkgs, ... }:
        with lib;
        let
          cfg = config.programs.unifi-threat-sync;
        in
        {
          options.programs.unifi-threat-sync = {
            enable = mkEnableOption "UniFi Threat Sync";

            package = mkOption {
              type = types.package;
              default = self.packages.${pkgs.system}.default;
              description = "The unifi-threat-sync package to use";
            };
          };

          config = mkIf cfg.enable {
            home.packages = [ cfg.package ];
          };
        };
    };
}

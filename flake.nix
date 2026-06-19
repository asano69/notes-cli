# notes-cli/flake.nix
{
  description = "Forked notes-cli";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-26.05";

  outputs = { self, nixpkgs }:
    let
      forAllSystems = nixpkgs.lib.genAttrs [ "x86_64-linux" "aarch64-linux" ];
      pkgsFor = system: import nixpkgs { inherit system; };
    in
    {
      # Used locally to mine vendorHash; the NixOS config does not consume this.
      packages = forAllSystems (system: {
        default = (pkgsFor system).buildGoModule {
          pname = "notes-cli";
          version = "0.0.4";
          src = self;
          vendorHash = "sha256-vQAy9veI7g+w9AyGqvtBWVGXLoVlnNF5/2YR3fecRzI=";
          meta = with (pkgsFor system).lib; {
            description = "Forked notes-cli";
            homepage = "https://github.com/asano69/notes-cli";
            license = licenses.mit;
          };
        };
      });

      # Kept for future use, if the root flake's inputs ever get edited.
      overlays.default = final: prev: {
        notes-cli = self.packages.${final.system}.default;
      };
    };
}

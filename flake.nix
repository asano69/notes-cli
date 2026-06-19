# notes-cli/flake.nix
{
  description = "Forked notes-cli";

  outputs = { self }: {
    overlays.default = final: prev: {
      notes-cli = final.buildGoModule {
        pname = "notes-cli";
        version = "0.0.1";
        src = self; # No hash needed: flake.lock on the consumer side tracks this content
        vendorHash = "sha256-vQAy9veI7g+w9AyGqvtBWVGXLoVlnNF5/2YR3fecRzI=";

        meta = with final.lib; {
          description = "Forked notes-cli";
          homepage = "https://github.com/asano69/notes-cli";
          license = licenses.mit;
        };
      };
    };
  };
}

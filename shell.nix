{ pkgs ? import <nixpkgs> {} }:

let
  # 使用したいパッケージを含んだPython環境を定義
  my-python = pkgs.python3.withPackages (ps: with ps; [
    pyyaml
    ruamel-yaml
    # requests  # 必要になったらコメントアウトを外すだけ
    # numpy
  ]);
in
pkgs.mkShell {
  buildInputs = [
    my-python
  ];
}

{
  lib,
  stdenv,
  buildGo127Module,
  ffmpeg,
  makeWrapper,
}: let
  version = "0.4.2";
in
  buildGo127Module {
    pname = "radiogogo";
    inherit version;

    src = lib.cleanSource ../.;

    vendorHash = "sha256-7j5BZx+meKZKW+w0vf0oGIg7/FB8KkooRMaJ8qi3qdU=";

    nativeBuildInputs = [makeWrapper];

    ldflags = [
      "-s"
      "-w"
    ];

    postInstall = ''
      wrapProgram $out/bin/radiogogo \
          --prefix PATH : ${lib.makeBinPath [ffmpeg]}
    '';

    meta = with lib; {
      homepage = "https://github.com/matteo=pacini/RadioGoGo";
      description = "Go-powered CLI to surf global radio waves via a sleek TUI";
      license = licenses.mit;
      maintainers = with maintainers; [matteopacini];
      mainProgram = "radiogogo";
    };
  }

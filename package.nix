{
  lib,
  buildGoModule,
}:

buildGoModule rec {
  pname = "ansipath";
  version = "0.2.0";

  src = ./.;

  vendorHash = null;
  ldflags = [
    "-s"
    "-w"
  ];

  meta = {
    description = "A colorized diagnostic view of shell PATH entries";
    homepage = "https://github.com/averyfreeman/ansipath";
    license = lib.licenses.gpl3Only;
    mainProgram = "ansipath";
    platforms = lib.platforms.unix;
  };
}

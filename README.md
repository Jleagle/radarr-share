# Library

  library:
    build: "library"
    container_name: "library"
    hostname: "library"
    restart: "unless-stopped"
    command: "--radarr-key=${LIBRARY_RADARR}"

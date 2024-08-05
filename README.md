# radarr-share

    services:
        radarr-share:
            image: "ghcr.io/jleagle/radarr-share:main"
            container_name: "radarr-share"
            hostname: "radarr-share"
            restart: "unless-stopped"
            entrypoint: "/root/radarr-share -radarr-key ${RADARR_KEY}"
            ports:
              - "7879:7879"

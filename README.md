# radarr-share

    services:
        radarr-share:
            image: "ghcr.io/jleagle/radarr-share:main"
            entrypoint: "/root/radarr-share -radarr-key ${RADARR_KEY}"
            ports:
              - "7879:7879"

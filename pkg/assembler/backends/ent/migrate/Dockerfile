FROM arigaio/atlas:latest-alpine@sha256:af71a72245d3ef2784ae87bcec9a9fd9a70f98544d1a14cc01bdd6485fa24805

WORKDIR /app

# Copy the migration files
COPY migrations ./migrations

# Copy the atlas migration script and make it executable
COPY atlas.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

ENTRYPOINT ["/app/entrypoint.sh"]

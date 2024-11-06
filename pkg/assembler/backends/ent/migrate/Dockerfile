FROM arigaio/atlas:latest-alpine@sha256:125f4b9d2a018fe9fd25b07922e7a6d3997b9bcb019070800fc506ed669e24da

WORKDIR /app

# Copy the migration files
COPY migrations ./migrations

# Copy the atlas migration script and make it executable
COPY atlas.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

ENTRYPOINT ["/app/entrypoint.sh"]

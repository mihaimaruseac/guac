FROM arigaio/atlas:latest-alpine@sha256:4491c802c7a6b5165c22c8d60ed7651ed5f683e00404808c6a271987f56f4d7c

WORKDIR /app

# Copy the migration files
COPY migrations ./migrations

# Copy the atlas migration script and make it executable
COPY atlas.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

ENTRYPOINT ["/app/entrypoint.sh"]

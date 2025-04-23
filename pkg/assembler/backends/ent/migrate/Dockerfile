FROM arigaio/atlas:latest-alpine@sha256:b65a6daaa060acfa5dedd6ce33fae1a1f328d598aebfaa11263e22e3aef9f50c

WORKDIR /app

# Copy the migration files
COPY migrations ./migrations

# Copy the atlas migration script and make it executable
COPY atlas.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

ENTRYPOINT ["/app/entrypoint.sh"]

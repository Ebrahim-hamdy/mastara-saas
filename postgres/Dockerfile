# Use the official PostgreSQL 16 Alpine image as our base.
FROM postgres:16.2-alpine

# Install the necessary build tools and git to compile the extension.
# We use a virtual package to easily clean up build dependencies later.
RUN apk add --no-cache --virtual .build-deps build-base postgresql-dev git && \
  \
  # Clone the pg_uuidv7 repository from GitHub.
  # Using a specific, stable tag is better than cloning the main branch.
  git clone --branch v1.4.0 https://github.com/fboulnois/pg_uuidv7.git /docker-entrypoint-initdb.d/pg_uuidv7 && \
  \
  # Change into the cloned directory.
  cd /docker-entrypoint-initdb.d/pg_uuidv7 && \
  \
  # Compile and install the extension into the PostgreSQL directory.
  make && make install && \
  \
  # Clean up the build dependencies to keep the final image lean.
  apk del .build-deps
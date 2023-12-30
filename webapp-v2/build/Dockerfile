# syntax = docker/dockerfile:1-experimental

###
# Build next.js app
###
FROM node:16.3

WORKDIR /build
run echo a
COPY webapp-v2/package.json webapp-v2/package-lock.json ./
RUN npm install --no-progress --loglevel http

COPY webapp-v2/ ./
RUN --mount=type=cache,target=/build/node_modules/.cache npm run build

ENTRYPOINT ["npm", "start"]

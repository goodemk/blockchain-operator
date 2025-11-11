FROM node:14-alpine AS builder

# Install git for npm dependencies that use git repositories
RUN apk add --no-cache git

WORKDIR /app

# ARG REPO_URL=https://github.com/goodemk/racecourse.git
# ARG REPO_BRANCH=master
# RUN git clone --depth 1 --branch ${REPO_BRANCH} ${REPO_URL} /app/source

COPY racecourse /app/source

WORKDIR /app/source/client
RUN npm install && npm cache clean --force
RUN npm run build

WORKDIR /app/source/server
RUN npm install --production && npm cache clean --force

FROM node:14-alpine

WORKDIR /app

COPY --from=builder /app/source/server ./server
COPY --from=builder /app/source/client/build ./client/build
COPY --from=builder /app/source/contracts ./contracts

EXPOSE 3000

WORKDIR /app/server

CMD ["node", "server.js"]

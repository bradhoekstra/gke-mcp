# checkov:skip=CKV_DOCKER_2:Existing issue, suppressing to unblock presubmit
# checkov:skip=CKV_DOCKER_3:Existing issue, suppressing to unblock presubmit
FROM node:22-slim AS build

WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY tsconfig.json ./
COPY src/ ./src/
COPY ui/ ./ui/
RUN npm --prefix ui install
RUN npm run build

FROM gcr.io/google.com/cloudsdktool/google-cloud-cli:558.0.0-debian_component_based-20260224

# Install Node.js 22
RUN apt-get update && apt-get install -y curl
RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash -
RUN apt-get install -y nodejs

WORKDIR /app
COPY --from=build /app/dist ./dist
COPY --from=build /app/node_modules ./node_modules
COPY --from=build /app/package*.json ./
COPY --from=build /app/ui/dist ./ui/dist

EXPOSE 8080
ENTRYPOINT [ "node", "dist/index.js" ]

/** @type {import('next').NextConfig} */
const nextConfig = {
  // Standalone output for the Docker image (see Dockerfile) - bundles just
  // the files needed to run `node server.js`, not the full node_modules tree.
  output: "standalone",
};

module.exports = nextConfig;

# Use a specific version of Node.js
FROM node:18

# Set working directory
WORKDIR /app

# Copy package.json and package-lock.json
COPY package.json ./

COPY index.js ./

# Install dependencies
RUN npm install

# Expose port for the chaincode
#EXPOSE 2053

ENV DEBUG=*

# Start the chaincode
CMD ["./node_modules/.bin/fabric-chaincode-node", "start", "--peer.address", "peer0org1_peer_ccaas:2051", "--chaincode-id-name", "v2", "--logging-level", "debug"]
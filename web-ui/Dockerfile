FROM node:22-alpine

WORKDIR /app

# Copy package files first to leverage Docker's cache layers
COPY package*.json ./

RUN npm install

COPY . .

CMD ["npm", "run", "dev"]
FROM node:16.14
WORKDIR /
COPY . .
RUN npm install
EXPOSE 50051
CMD ["node","grpc_server.js"]

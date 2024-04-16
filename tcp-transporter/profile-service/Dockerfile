FROM node:20-alpine
ENV NODE_ENV production
WORKDIR / 
COPY ./*.js ./
COPY ./*.json ./
RUN npm install
CMD [ "node", "/profile.js" ]
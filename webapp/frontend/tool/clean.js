const path = require('path');
const rimraf = require('rimraf');

rimraf.sync(`${__dirname}/../../public/**/{*.png,*.json,*.js,*.html,static/**}`);

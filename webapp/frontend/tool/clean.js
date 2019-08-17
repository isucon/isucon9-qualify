const rimraf = require('rimraf');

rimraf('../public/**/!(upload/**)', {}, () => {});

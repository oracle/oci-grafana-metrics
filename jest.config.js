const config = {
  verbose: true,
};
  
module.exports = config;
  
// Or async function
module.exports = async () => {
  return {
    verbose: true,
    rootDir: process.cwd(),
  };
};
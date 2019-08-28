const express = require('express');
const router = express.Router();

// Send a happy 200 and returncode:Success if we are running properly.
router.get('/', (req, res) => {
  res.type('json').send({returnCode: "Success"});
});

module.exports = router;

// Super unsafe stuff just to work with the temp system. REMOVE SOON.
process.env['NODE_TLS_REJECT_UNAUTHORIZED'] = '0';

// Adding dependencies.
const express = require('express');
const path = require('path');
const bodyParser = require('body-parser');
const rp = require('request-promise-native');


// Adding the code for the routes.
const healthcheck = require('./routes/healthcheck');

// Express server is "app"
const app = express();

// Disabling the x-powered-by: Express header, for security.
app.disable('x-powered-by');

// ***** Middleware to allow CORS access.
app.use((req, res, next) => {
  res.setHeader('Access-Control-Allow-Origin', '*');
  res.setHeader('Access-Control-Allow-Methods', 'GET, POST, OPTIONS, PUT, DELETE');
  res.setHeader('Access-Control-Allow-Headers', 'X-Requested-With,Authorization,X-PINGOTHER,Content-Type');
  // Check if this is a preflight request. If so, send 200. Otherwise, pass it forward.
  if (req.method === 'OPTIONS') {
      //respond with 200
      res.sendStatus(200);
  } else {
    next();
  }
});

// Body parser and static file directories.
app.use(bodyParser.json());
app.use(express.static(path.join(__dirname, 'public')));

// Plug in the main routes.
app.use('/healthcheck', healthcheck);

app.post('/tripsCheck', async (req, res, next) => {
  const channels = req.body.channels;
  const queryURL = `${req.body.url}/bcsgw/rest/v1/transaction/query`;
  const chaincode = req.body.chaincode;
  // Boolean so we don't double-send a reply if there's an error.
  let markedError = false;
  // First, we create an empty object that will hold our results...
  let allVehiclesAllTrips = {};
  // Now, let's go through every channel (with a for loop so we don't have to deal with setting a HOF to async).
  for (let i=0; i<channels.length; i++) {
    // Setting the request options for this vendor.
    let requestOptions = {
      method: 'POST',
      uri: queryURL,
      json: true,
      body: {
        channel : channels[i],
        chaincode : chaincode,
        method: "getAllTripsAllVehicles",
        args: []
      }
    };
    if (req.body.restUsername && req.body.restPassword) {
      authString = new Buffer(req.body.restUsername + ':' + req.body.restPassword).toString('base64');
      requestOptions.headers = {
        Authorization : `Basic ${authString}`
      };
    }
    console.log(`We are starting a request on the channel ${channels[i]}`);
    console.log(`===========REQUEST OPTIONS==============`);
    console.log(requestOptions);
    console.log('========================================');
    try {
      // Await the request to get its response.
      let response = await rp(requestOptions);
      // Check the returnCode of the response, make sure it's a success.
      if (response.returnCode !== 'Success') {
        console.error('Return code was not success. ', response);
        if (!markedError) {
          res.status(500).send(response.info);
          markedError = true;
        }
      } else { // If the return code is a success...
        console.log(`Response received making call on the channel ${channels[i]}`);
        // console.log(`This is the response - `, response);
        // Parse the JSON string in the result and put it into the object
        if (response.result.payload) {
          console.log(`=========response.result.payload==========`);
          console.log(response.result.payload);
          console.log(`=====================================`);
          allVehiclesAllTrips[channels[i]] = JSON.parse(response.result.payload);
        } else {
          allVehiclesAllTrips[channels[i]] = JSON.parse(response.result);
        }
        console.log('Here is the vehicleTripHistory -');
        console.log(allVehiclesAllTrips);
      }
    } catch (err) {
      console.error(`Error received making call on the channel ${channels[i]}`);
      console.error(err);
      if (!markedError) {
        res.status(500).send(err);
        markedError = true;
      }
    }
  }
  if (!markedError) {
    res.send(allVehiclesAllTrips);
  }
});


//====== Route for getting the list of registered vehicles and their state =====
app.post('/vehiclesCheck', async (req, res, next) => {
  const queryURL = `${req.body.url}/bcsgw/rest/v1/transaction/query`;
  const channels = req.body.channels;
  const chaincode = req.body.chaincode;
  let allVehicles = {};
  for (let i=0; i< channels.length; i++) {
    let requestOptions = {
        method: 'POST',
        uri: queryURL,
        json: true,
        body: {
          channel : channels[i],
          chaincode : chaincode,
          method: "seeAllVehicles",
          args: []
        }
      };
      if (req.body.restUsername && req.body.restPassword) {
        authString = new Buffer(req.body.restUsername + ':' + req.body.restPassword).toString('base64');
        requestOptions.headers = {
          Authorization : `Basic ${authString}`
        };
      }
      console.log(`===========VEHICLESCHECK REQUEST OPTIONS==============`);
      console.log(requestOptions);
      console.log('========================================');  
      try {
          // Await the request to get its response.
          let response = await rp(requestOptions);
          // Check the returnCode of the response, make sure it's a success.
          if (response.returnCode !== 'Success') {
            console.log('THERE WAS A PROBLEM.');
            console.log(response);
            res.status(500).send(response.info);
            return;
          } else { // If the return code is a success...
            console.log(`Response received making call on the channel ${channels[i]}`);
            // Parse the JSON string in the result and put it into the object
            if (response.result.payload) {
              allVehicles[channels[i]] = JSON.parse(response.result.payload);
            } else {
              allVehicles[channels[i]] = JSON.parse(response.result);
            }
            console.log('Here are the channelVehicles -');
            console.log(allVehicles[channels[i]]);
          }
      } catch (err) {
        console.error(`Error received making call on the channel ${channels[i]}`);
        console.error(err);
        res.status(500).send(err);
        next();
      }
    }
    console.log('Here is allVehicles - ');
    console.log(allVehicles);
    res.send(allVehicles);
});


app.post('/registerVehicle', async (req, res, next) => {
  const queryURL = `${req.body.url}/bcsgw/rest/v1/transaction/invocation`;
  const channels = req.body.channels;
  const chaincode = req.body.chaincode;
  const vehicleSerial = req.body.serial;
  const vehicleOwner = req.body.owner;
  const vehicleType = req.body.type;

  let requestOptions = {
    method: 'POST',
    uri: queryURL,
    json: true,
    body: {
      channel : channels[0],
      chaincode : chaincode,
      method: "registerVehicle",
      args: [vehicleSerial, vehicleOwner, vehicleType]
    }
  };
  if (req.body.restUsername && req.body.restPassword) {
    authString = new Buffer(req.body.restUsername + ':' + req.body.restPassword).toString('base64');
    requestOptions.headers = {
      Authorization : `Basic ${authString}`
    };
  }
  console.log(`===========REQUEST OPTIONS==============`);
  console.log(requestOptions);
  console.log('========================================');
  try {
      // Await the request to get its response.
      let response = await rp(requestOptions);
      // Check the returnCode of the response, make sure it's a success.
      if (response.returnCode !== 'Success') {
        console.log('THERE WAS A PROBLEM.');
        console.log(response);
        res.status(500).send(response.info);
        return;
      } else { // If the return code is a success...
        console.log('Add vehicle success.');
        console.log(response);
        res.sendStatus(200);
      }
  } catch (err) {
    console.error(`Error received making call`);
    console.error(err);
    res.status(500).send(err);
  }
});


app.post('/takeOutOfService', async (req, res, next) => {
  const queryURL = `${req.body.url}/bcsgw/rest/v1/transaction/invocation`;
  const channels = req.body.channels;
  const chaincode = req.body.chaincode;
  const vehicleSerial = req.body.serial;
  let requestOptions = {
    method: 'POST',
    uri: queryURL,
    json: true,
    body: {
      channel : channels[0],
      chaincode : chaincode,
      method: "takeOutOfService",
      args: [vehicleSerial]
    }
  };
  if (req.body.restUsername && req.body.restPassword) {
    authString = new Buffer(req.body.restUsername + ':' + req.body.restPassword).toString('base64');
    requestOptions.headers = {
      Authorization : `Basic ${authString}`
    };
  }
  console.log(`===========REQUEST OPTIONS==============`);
  console.log(requestOptions);
  console.log('========================================');
  try {
      // Await the request to get its response.
      let response = await rp(requestOptions);
      // Check the returnCode of the response, make sure it's a success.
      if (response.returnCode !== 'Success') {
        console.log('THERE WAS A PROBLEM.');
        console.log(response);
        res.status(500).send(response.info);
        return;
      } else { // If the return code is a success...
        console.log('Take out of service success.');
        console.log(response);
        res.sendStatus(200);
      }
  } catch (err) {
    console.error(`Error received making call`);
    console.error(err);
    res.status(500).send(err);
  }
});


app.post('/putInService', async (req, res, next) => {
  const queryURL = `${req.body.url}/bcsgw/rest/v1/transaction/invocation`;
  const channels = req.body.channels;
  const chaincode = req.body.chaincode;
  const vehicleSerial = req.body.serial;
  const currentLocation = req.body.location;
  let requestOptions = {
    method: 'POST',
    uri: queryURL,
    json: true,
    body: {
      channel : channels[0],
      chaincode : chaincode,
      method: "putInService",
      args: [vehicleSerial,currentLocation]
    }
  };
  if (req.body.restUsername && req.body.restPassword) {
    authString = new Buffer(req.body.restUsername + ':' + req.body.restPassword).toString('base64');
    requestOptions.headers = {
      Authorization : `Basic ${authString}`
    };
  }
  console.log(`===========REQUEST OPTIONS==============`);
  console.log(requestOptions);
  console.log('========================================');
  try {
      // Await the request to get its response.
      let response = await rp(requestOptions);
      // Check the returnCode of the response, make sure it's a success.
      if (response.returnCode !== 'Success') {
        console.log('THERE WAS A PROBLEM.');
        console.log(response);
        res.status(500).send(response.info);
        return;
      } else { // If the return code is a success...
        console.log('Take out of service success.');
        console.log(response);
        res.sendStatus(200);
      }
  } catch (err) {
    console.error(`Error received making call`);
    console.error(err);
    res.status(500).send(err);
  }
});



app.post('/takeTrip', async (req, res, next) => {
  const queryURL = `${req.body.url}/bcsgw/rest/v1/transaction/invocation`;
  const channels = req.body.channels;
  const vehicleChaincode = req.body.vehicleChaincode;
  const tripChaincode = req.body.tripChaincode;
  const vehicleSerial = req.body.serial;
  const startingLocation = req.body.location.split(',');
  const tripLength = req.body.length;

  const startLat = startingLocation[0];
  const startLong = startingLocation[1];

  let distanceChangeLat = Math.floor(Math.random()*tripLength*1000);
  let distanceChangeLong = Math.floor(Math.random()*tripLength*1000);

  // Randomly choosing between adding or subtracting to location depending on if the change integer is even or odd.
  if (distanceChangeLong%2 === 0) {
    distanceChangeLat = -1*(distanceChangeLat/1000000);
  } else {
    distanceChangeLat = distanceChangeLat/1000000;
  }
  // Randomly choosing between adding or subtracting to location depending on if the change integer is even or odd. - Longitudinal version
  if (distanceChangeLat%2 === 0) {
    distanceChangeLong = -1*(distanceChangeLong/1000000);
  } else {
    distanceChangeLong = distanceChangeLong/1000000;
  }

  const endLat = parseFloat(startLat) + distanceChangeLat;
  const endLong = parseFloat(startLong) + distanceChangeLong;
  const startTime = `${Date.now()}`;
  const endTime = `${Date.now() + (tripLength * 60000)}`;

  let tripRequestOptions = {
    method: 'POST',
    uri: queryURL,
    json: true,
    body: {
      channel : channels[0],
      chaincode : tripChaincode,
      method: "recordTrip",
      args: [vehicleSerial, startTime, startLat, startLong, endTime, endLat.toFixed(6), endLong.toFixed(6)]
    }
  };
  if (req.body.restUsername && req.body.restPassword) {
    authString = new Buffer(req.body.restUsername + ':' + req.body.restPassword).toString('base64');
    tripRequestOptions.headers = {
      Authorization : `Basic ${authString}`
    };
  }
  console.log(`===========TRIP REQUEST OPTIONS==============`);
  console.log(tripRequestOptions);
  console.log('========================================');


  const newCurrentLocation = `${endLat.toFixed(6)},${endLong.toFixed(6)}`;

  let updateLocationRequestOptions = {
    method: 'POST',
    uri: queryURL,
    json: true,
    body: {
      channel : channels[0],
      chaincode : vehicleChaincode,
      method: "updateLocation",
      args: [vehicleSerial, newCurrentLocation]
    }
  };

  if (req.body.restUsername && req.body.restPassword) {
    authString = new Buffer(req.body.restUsername + ':' + req.body.restPassword).toString('base64');
    updateLocationRequestOptions.headers = {
      Authorization : `Basic ${authString}`
    };
  }

  try {
    // Await the request to get its response.
    let response = await rp(tripRequestOptions);
    // Check the returnCode of the response, make sure it's a success.
    if (response.returnCode !== 'Success') {
      console.log('THERE WAS A PROBLEM LOGGING TRIP.');
      console.log(response);
      res.status(500).send(response.info);
      next();
    }
    console.log(`===========UPDATE LOCATION REQUEST OPTIONS==============`);
    console.log(updateLocationRequestOptions);
    console.log('========================================');
      let response2 = await rp(updateLocationRequestOptions);
    if (response2.returnCode !== 'Success') {
      console.log('THERE WAS A PROBLEM UPDATING CURRENT LOCATION. IT IS OUT OF  SYNC.');
      console.log(response2);
      res.status(500).send(response2.info);
      next();
    }
    console.log('Update complete.');
    console.log(response);
    console.log(response2);
    res.status(200).send({newLocation:newCurrentLocation});
  } catch (err) {
    console.error(`Error received making call`);
    console.error(err);
    res.status(500).send(err);
  }
});


// Catch 404 and forward to error handler
app.use((req, res, next) => {
  var err = new Error('Not Found');
  err.status = 404;
  next(err);
});

// error handler
app.use((err, req, res, next) => {
  // set locals, only providing error in development
  res.locals.message = err.message;
  res.locals.error = req.app.get('env') === 'development' ? err : {};
  // render the error page
  res.status(err.status || 500);
  res.send(err);
});

module.exports = app;

/**
 * @license
 * Copyright (c) 2014, 2018, Oracle and/or its affiliates.
 * The Universal Permissive License (UPL), Version 1.0
 */
define(['ojs/ojcore', 'knockout', 'jquery', 'appController', 'ojs/ojarraydataprovider', 'ojs/ojknockout', 'ojs/ojlistview', 'ojs/ojselectcombobox', 'ojs/ojbutton'],
  function(oj, ko, $, app, ArrayDataProvider) {
    function ScooterCoViewModel() {
      const self = this;

//============= Functions we need for this to work. ============================

/////////// This function makes the form to register vehicle visible. //////////
      self.registerVehicle = function() {
        self.vehicleRegisterAttempt(!self.vehicleRegisterAttempt());
      };

/////// This function registers a new vehicle on the blockchain. ///////////////
      self.makeVehicle = function(event, current, bindingContext) {
        event.target.disabled = true;
        if (self.serial() && self.vehicleType()) {
          const addVehicleRequestOptions = {
            channels : [app.scooterCoChannel],
            url : app.scooterCoRestProxy,
            chaincode : app.vehicleChaincode,
            serial : self.serial(),
            type : self.vehicleType(),
            owner : "ScooterCo"
          };
          if (app.cloudChain) {
            addVehicleRequestOptions.restUsername = app.restUsername;
            addVehicleRequestOptions.restPassword = app.restPassword;
          }
          $.ajax({
            type: 'POST',
            url: 'http://localhost:3000/registerVehicle',
            data: JSON.stringify(addVehicleRequestOptions),
            contentType: 'application/json',
            success: (data) => {
              self.dataArray.push({
                Key : self.serial(),
                Record: {
                  CurrentLocation: "oos",
                  VehicleType: self.vehicleType(),
                  InService : false
                }
              });
              self.vehicleRegisterAttempt(false);
              alert('Vehicle Registered Successfully!');
              event.target.disabled = false;
            },
            failure: (jqXHR, textStatus, errorThrown) => {
              alert('AJAX failure');
              event.target.disabled = false;
            },
            complete: (jqXHR) => {
              if (jqXHR.status !== 200) {
                alert(`There was an error with your request.\n${jqXHR.responseText}`);
                event.target.disabled = false;
              }
            }
          });
        } else {
          alert('Please fill out both fields before completing registration.');
          event.target.disabled = true;
        }
      }

///// This function takes an out-of-service vehicle and puts it in service /////
      self.putVehicleInService = function(event, current, bindingContext) {
        event.target.disabled = true;
        const serial = current.data.Key;
        const location = '38.891885,-77.036432'
        const putInServiceRequestOptions = {
          channels : [app.scooterCoChannel],
          url : app.scooterCoRestProxy,
          chaincode : app.vehicleChaincode,
          serial : serial,
          location : location
        };
        if (app.cloudChain) {
          putInServiceRequestOptions.restUsername = app.restUsername;
          putInServiceRequestOptions.restPassword = app.restPassword;
        }
        $.ajax({
          type: 'POST',
          url: 'http://localhost:3000/putInService',
          data: JSON.stringify(putInServiceRequestOptions),
          contentType: 'application/json',
          success: (data) => {
            current.data.Record.InService = true;
            current.data.Record.CurrentLocation = location;
            self.dataArray.replace(current.data,current.data);
            event.target.disabled = false;
          },
          failure: (jqXHR, textStatus, errorThrown) => {
            alert('AJAX failure');
            event.target.disabled = false;
          },
          complete: (jqXHR) => {
            if (jqXHR.status !== 200) {
              alert(`There was an error with your request.\n${jqXHR.responseText}`);
              event.target.disabled = false;
            }
          }
        });
      }

///// This function makes it possible for the trip record items to appear. /////
      self.recordTrip = function(event, current, bindingContext) {
        current.data.thisExists = true;
        self.dataArray.replace(current.data,current.data);
      }

////////////////// This completes the trip. ////////////////////////////////////
      self.completeRecordTrip = function(event, current, bindingContext) {
        event.target.disabled = true;
        const createTripRequestOptions = {
          channels : [app.scooterCoChannel],
          url : app.scooterCoRestProxy,
          vehicleChaincode : app.vehicleChaincode,
          tripChaincode : app.tripChaincode,
          serial : current.data.Key,
          location : current.data.Record.CurrentLocation,
          length : parseInt(self.tripLengthVal())
        };
        if (app.cloudChain) {
          createTripRequestOptions.restUsername = app.restUsername;
          createTripRequestOptions.restPassword = app.restPassword;
        }
        $.ajax({
          type: 'POST',
          url: 'http://localhost:3000/takeTrip',
          data: JSON.stringify(createTripRequestOptions),
          contentType: 'application/json',
          success: (data) => {
           current.data.Record.CurrentLocation = data.newLocation;
           current.data.thisExists = false;
           self.dataArray.replace(current.data,current.data);
           event.target.disabled = false;
          },
          failure: (jqXHR, textStatus, errorThrown) => {
            alert('AJAX failure');
            event.target.disabled = false;
          },
          complete: (jqXHR) => {
            if (jqXHR.status !== 200) {
              alert(`There was an error with your request.\n${jqXHR.responseText}`);
              event.target.disabled = false;
            }
          }
        });
      }

/////////////// This function takes a vehicle out of service. //////////////////
      self.takeOutOfService = function(event, current, bindingContext) {
        event.target.disabled = true;
        const serial = current.data.Key;
        const takeOutOfServiceRequestOptions = {
          channels : [app.scooterCoChannel],
          url : app.scooterCoRestProxy,
          chaincode : app.vehicleChaincode,
          serial : serial
        };
        if (app.cloudChain) {
          takeOutOfServiceRequestOptions.restUsername = app.restUsername;
          takeOutOfServiceRequestOptions.restPassword = app.restPassword;
        }
        $.ajax({
          type: 'POST',
          url: 'http://localhost:3000/takeOutOfService',
          data: JSON.stringify(takeOutOfServiceRequestOptions),
          contentType: 'application/json',
          success: (data) => {
            current.data.Record.InService = false;
            self.dataArray.replace(current.data,current.data);
            event.target.disabled = false;
          },
          failure: (jqXHR, textStatus, errorThrown) => {
            alert('AJAX failure');
            event.target.disabled = false;
          },
          complete: (jqXHR) => {
            if (jqXHR.status !== 200) {
              alert(`There was an error with your request.\n${jqXHR.responseText}`);
              event.target.disabled = false;
            }
          }
        });
      }

/////////// This function takes the vehicle data & adds it to the //////////////
/////////// observableArray that feeds the dataprovider. ///////////////////////
      function makeVehicleList(data) {
        self.dataArray([]);
        data.forEach((datum) => {
          self.dataArray.push(datum);
        });
      }

///////////// All the ko.observables we'll be using in this. ///////////////////
      self.serial = ko.observable('');
      self.vehicleType = ko.observable('');
      self.tripLengthVal = ko.observable('');
      self.vehicleRegisterAttempt = ko.observable(false);
      self.dataArray = ko.observableArray([]);
      self.dataProvider = new ko.observable(new ArrayDataProvider(self.dataArray, {keyAttributes: 'Key'}));


//================== Initial AJAX request area. ================================

      const requestOptions = {
        channels : [app.scooterCoChannel],
        url : app.scooterCoRestProxy,
        chaincode : app.vehicleChaincode
      };
      if (app.cloudChain) {
        requestOptions.restUsername = app.restUsername;
        requestOptions.restPassword = app.restPassword;
      };
      self.connected = function() {
        $.ajax({
          type: 'POST',
          url: 'http://localhost:3000/vehiclesCheck',
          data: JSON.stringify(requestOptions),
          contentType: 'application/json',
          success: (data) => {
            if (data[app.scooterCoChannel]) {
              makeVehicleList(data[app.scooterCoChannel])
            } else {
              alert('Currently no vehicles for this vendor.')
              makeVehicleList([]);
            }
          },
          failure: (jqXHR, textStatus, errorThrown) => {
            alert('AJAX failure');
          },
          complete: (jqXHR) => {
            if (jqXHR.status !== 200) {
              alert(`There was an error with your request.\n${jqXHR.responseText}`);
            }
          }
        });
      };
      self.disconnected = function() {
        // Implement if needed
      };
      self.transitionCompleted = function() {
        // Implement if needed
      };
    }
    return new ScooterCoViewModel();
  }
);

/**
 * @license
 * Copyright (c) 2014, 2018, Oracle and/or its affiliates.
 * The Universal Permissive License (UPL), Version 1.0
 */
/*
 * Your dashboard ViewModel code goes here
 */
define(['ojs/ojcore', 'knockout', 'jquery', 'appController', 'ojs/ojknockout', 'ojs/ojbutton', 'ojs/ojinputtext'],
 function(oj, ko, $, app) {

    function LoginViewModel() {
      var self = this;

      self.signedIn = app.signedIn;
      self.cityRestProxy = ko.observable("");
      self.riderCoRESTProxy = ko.observable("");
      self.scooterCoRestProxy = ko.observable("");
      self.riderCoChannel = ko.observable("");
      self.scooterCoChannel = ko.observable("");
      self.vehicleChaincode = ko.observable("");
      self.tripChaincode = ko.observable("");
      self.restUsername = ko.observable("");
      self.restPassword = ko.observable("");
      self.cloudChain = ko.observable(false);

      self.submitData = function() {
        if (self.cloudChain()) {
          if (self.restUsername() && self.restPassword()) {
            app.cityRestProxy = self.cityRestProxy();
            app.riderCoRESTProxy = self.riderCoRESTProxy();
            app.scooterCoRestProxy = self.scooterCoRestProxy();
            app.riderCoChannel = self.riderCoChannel();
            app.scooterCoChannel = self.scooterCoChannel();
            app.vehicleChaincode = self.vehicleChaincode();
            app.tripChaincode = self.tripChaincode();
            app.restUsername = self.restUsername();
            app.restPassword = self.restPassword();
            app.cloudChain = self.cloudChain();
            self.signedIn(true);
            app.signedIn(true);
            app.pageid = 'city';
            app.router.go('city');
            app.refreshMenu();
          } else {
            alert('Please enter both the username and password for REST proxy access, or uncheck the Blockchain Cloud Service option.');
          }
        } else {
          app.cityRestProxy = self.cityRestProxy();
          app.riderCoRESTProxy = self.riderCoRESTProxy();
          app.scooterCoRestProxy = self.scooterCoRestProxy();
          app.riderCoChannel = self.riderCoChannel();
          app.scooterCoChannel = self.scooterCoChannel();
          app.vehicleChaincode = self.vehicleChaincode();
          app.tripChaincode = self.tripChaincode();
          app.restUsername = self.restUsername();
          app.restPassword = self.restPassword();
          app.cloudChain = self.cloudChain();
          self.signedIn(true);
          app.signedIn(true);
          app.pageid = 'city';
          app.router.go('city');
          app.refreshMenu();
        }
      }


      self.connected = function() {
        app.refreshMenu();
      };
      self.disconnected = function() {
        // Implement if needed
      };
      self.transitionCompleted = function() {
        // Implement if needed
      };
    }
    return new LoginViewModel();
  }
);

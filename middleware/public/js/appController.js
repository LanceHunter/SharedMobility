/**
 * @license
 * Copyright (c) 2014, 2018, Oracle and/or its affiliates.
 * The Universal Permissive License (UPL), Version 1.0
 */
/*
 * Your application specific code will go here
 */
define(['ojs/ojcore', 'knockout', 'ojs/ojmodule-element-utils', 'ojs/ojmodule-element', 'ojs/ojrouter', 'ojs/ojknockout', 'ojs/ojarraytabledatasource'],
  function(oj, ko, moduleUtils) {
    function ControllerViewModel() {
      const self = this;
      self.userLogin = ko.observable("");
      self.pageid = 'login';
      self.signedIn = ko.observable(false);

      // Variables used for the requests...
      self.cityRestProxy = "";
      self.riderCoRESTProxy = "";
//      self.scooterCoRestProxy = "";
      self.riderCoChannel = "";
//      self.scooterCoChannel = "";
      self.vehicleChaincode = "";
      self.tripChaincode = "";
      self.cloudChain = false;
      self.restUsername = "customertenant@oracle.com";
      self.restPassword = "Welcome1";


      self.reset = function() {
        self.router.go('login');
        location.reload();
      }

      // Media queries for repsonsive layouts
      const smQuery = oj.ResponsiveUtils.getFrameworkQuery(oj.ResponsiveUtils.FRAMEWORK_QUERY_KEY.SM_ONLY);
      self.smScreen = oj.ResponsiveKnockoutUtils.createMediaQueryObservable(smQuery);

       // Router setup
       self.router = oj.Router.rootInstance;
       self.router.configure({
         'login': {label: 'login', isDefault: true},
         'city': {label: 'City'},
         'riderco': {label: 'RiderCo'} //,
//         'scooterco': {label: 'ScooterCo'}
       });
      oj.Router.defaults['urlAdapter'] = new oj.Router.urlParamAdapter();

      self.moduleConfig = ko.observable({'view':[], 'viewModel':null});

      self.loadModule = function() {
        ko.computed(function() {
          var name = self.router.moduleConfig.name();
          var viewPath = 'views/' + name + '.html';
          var modelPath = 'viewModels/' + name;
          var masterPromise = Promise.all([
            moduleUtils.createView({'viewPath':viewPath}),
            moduleUtils.createViewModel({'viewModelPath':modelPath})
          ]);
          masterPromise.then(
            function(values){
              self.moduleConfig({'view':values[0],'viewModel':values[1]});
            }
          );
        });
      };

      // Navigation setup
      var navData = [
      {name: 'City', id: 'city',
       iconClass: 'oj-navigationlist-item-icon demo-icon-font-24 demo-home-icon-24'},
      {name: 'RiderCo', id: 'riderco',
       iconClass: 'oj-navigationlist-item-icon demo-icon-font-24 demo-people-icon-24'} //,
      // {name: 'ScooterCo', id: 'scooterco',
      //  iconClass: 'oj-navigationlist-item-icon demo-icon-font-24 demo-people-icon-24'}
      ];
      self.navDataSource = new oj.ArrayTableDataSource(navData, {idAttribute: 'id'});


      self.refreshMenu = function () {
        if (self.pageid === 'login') {
          document.getElementById("header").style.visibility = 'hidden';
          document.getElementById("footer").style.visibility = 'hidden';
        } else {
          document.getElementById("header").style.visibility = 'visible';
          document.getElementById("footer").style.visibility = 'visible';
        }
      }


      // Header
      // Application Name used in Branding Area
      self.appName = ko.observable("Shared Mobility Device Services");
      // User Info used in Global Navigation area
      self.userLogin = ko.observable("");

      // Footer
      function footerLink(name, id, linkTarget) {
        this.name = name;
        this.linkId = id;
        this.linkTarget = linkTarget;
      }
      self.footerLinks = ko.observableArray([
        new footerLink('About Oracle', 'aboutOracle', 'http://www.oracle.com/us/corporate/index.html#menu-about'),
        new footerLink('Contact Us', 'contactUs', 'http://www.oracle.com/us/corporate/contact/index.html'),
        new footerLink('Legal Notices', 'legalNotices', 'http://www.oracle.com/us/legal/index.html'),
        new footerLink('Terms Of Use', 'termsOfUse', 'http://www.oracle.com/us/legal/terms/index.html'),
        new footerLink('Your Privacy Rights', 'yourPrivacyRights', 'http://www.oracle.com/us/legal/privacy/index.html')
      ]);
     }

     return new ControllerViewModel();
  }
);

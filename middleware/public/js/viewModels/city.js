/**
 * @license
 * Copyright (c) 2014, 2018, Oracle and/or its affiliates.
 * The Universal Permissive License (UPL), Version 1.0
 */
define(['ojs/ojcore', 'knockout', 'jquery', 'appController', 'ojs/ojknockout', 'ojs/ojbutton', 'ojs/ojchart'], function(oj, ko, $, app) {

  function CityViewModel() {
    var self = this;

//========== Beginning of the area used to make the usage pie chart ============
    function makePieChart(data) {
      Object.keys(data).forEach((key) => {
        data[key].forEach((vehicle) => {
          vehicle.Record.forEach((trip) => {
            let tripDate = new Date(trip.Timestamp);
            let tripDateString = `${tripDate.getMonth()+1}-${tripDate.getDate()}-${tripDate.getFullYear()}`
            if (!totalTrips[tripDateString]) {
              self.dateOptions.push(tripDateString);
              if (self.dateOptions.indexOf(tripDateString) < 0) {
                self.dateOptions.push(tripDateString);
              }
              totalTrips[tripDateString] = 1;
            } else {
              totalTrips[tripDateString]++
            }
            if (!pieSeries[tripDateString]) {
              pieSeries[tripDateString] = [
                {name: "< 15 minutes", items: [0]},
                {name: "15 - 29 minutes", items: [0]},
                {name: "30 - 44 minutes", items: [0]},
                {name: "45 - 60 minutes", items: [0]},
                {name: "> 60 minutes", items: [0]}
              ];
            }
            totalTrips.All++;
            let duration = parseInt(trip.Value.EndTime) - parseInt(trip.Value.StartTime);
            if (duration < 900001) {
              pieSeries.All[0].items[0]++
              pieSeries[tripDateString][0].items[0]++
            } else if (duration < 1799999) {
              pieSeries.All[1].items[0]++
              pieSeries[tripDateString][1].items[0]++
            } else if (duration < 2699999) {
              pieSeries.All[2].items[0]++
              pieSeries[tripDateString][2].items[0]++
            } else if (duration < 3600000) {
              pieSeries.All[3].items[0]++
              pieSeries[tripDateString][3].items[0]++
            } else {
              pieSeries.All[4].items[0]++
              pieSeries[tripDateString][4].items[0]++
            }
          });
        });
      });
      self.pieSeriesValue(pieSeries.All);
      self.totalTrips(totalTrips.All);
    }

    // self.dateOptions = ko.observableArray(['All']);

    self.changeDate1 = function() {
      let dateChosen = $('#durationDate').val();
      self.pieSeriesValue(pieSeries[dateChosen]);
      self.totalTrips(totalTrips[dateChosen]);
    }

    let totalTrips = {
      All : 0
    };

    /* chart data */
    let pieSeries = {
      All : [
        {name: "< 15 minutes", items: [0]},
        {name: "15 - 29 minutes", items: [0]},
        {name: "30 - 44 minutes", items: [0]},
        {name: "45 - 60 minutes", items: [0]},
        {name: "> 60 minutes", items: [0]}
      ]
    };

    self.totalTrips = ko.observable(totalTrips);
    self.pieSeriesValue = ko.observableArray(pieSeries.All);
//================ End of the usage pie chart area. ============================


//========== Beginning of the bar graph for vehicles in service area ===========
    self.currentDate = new Date().toString();
    const isInService = ({ Value: { InService }}) => InService;
    const count = (pred, xs) => xs.reduce((sum, x) => sum + pred(x), 0);
    self.stackValue = ko.observable('on');
    self.stackLabelValue = ko.observable('on');
    self.orientationValue = ko.observable('vertical');

    const data = ko.observableArray([]);

    const graphData = ko.computed(() => {
      const inService = count(isInService, data());
      return [
        { name: "In Service", items: [inService]},
        { name: "Out of Service", items: [data().length - inService]}
      ];
    });

    let barSeries = [
      {
        name: "In-Service",
        items:[0, 0]
      },
      {
        name: "Out-of-Service",
        items:[0, 0]
      }
    ];

    const barGroups = ["RiderCo", "ScooterCo"];


    function useVehicleData(data) {
        let i = 0;
        for (const channel in data) {
            if (data.hasOwnProperty(channel)) {
                data[channel].forEach(element => {
                    if (element.Record.InService) {
                        barSeries[0].items[i]++;
                    } else {
                        barSeries[1].items[i]++;
                    }
                });
            }
            i++;
        }
        self.currentDate = new Date().toString();
        self.barSeriesValue(barSeries);
    };

    self.barSeriesValue = ko.observableArray(barSeries);
    self.barGroupsValue = ko.observableArray(barGroups);
//============ End of the bar graph for vehicles in service area ===============


//============= Area for the line graph of trips by hour =======================
    self.useDate = new Date();
//    self.dateOptions = ko.observableArray(['All']);

    const vendor = {
      "testvendor.channel": "RiderCo",
      "secondtestvendor.channel": "ScooterCo"
    };
    const groups = ["12 AM", "1 AM", "2 AM", "3 AM", "4 AM", "5 AM", "6 AM", "7 AM", "8 AM", "9 AM", "10 AM", "11 AM", "12 PM", "1 PM", "2 PM", "3 PM", "4 PM", "5 PM", "6 PM", "7 PM", "8 PM", "9 PM", "10 PM", "11 PM"];

    let series = {
      All : []
    };

    function getEmptyItems() {
      let items = new Array(24);
      items.fill(0);
      return items;
    }

    function getSeriesData(vendorData, channelSeriesData) {
      vendorData.forEach(vehicleData => {
        let vehicleTripLogs = vehicleData.Record;
        vehicleTripLogs.forEach(trip => {
          channelSeriesData.items[getHours(trip)] += 1 ;
        })
      });
    }


    function getSeriesDataForDate(listOfTrips, channelSeriesData) {
      listOfTrips.forEach((trip) => {
        channelSeriesData.items[getHours(trip)] += 1 ;
      });
    }


    function getHours(trip) {
      let date = new Date(parseInt(trip.Value.StartTime));
      return date.getHours();
    }


    function filterByDate(scooters) {
      let tripsByDate = {};
      scooters.forEach((scooter) => {
        scooter.Record.forEach((trip) => {
          let tripDate = new Date(trip.Timestamp);
          let tripDateString = `${tripDate.getMonth()+1}-${tripDate.getDate()}-${tripDate.getFullYear()}`;
          if (tripsByDate[tripDateString]) {
            tripsByDate[tripDateString].push(trip);
          } else {
            tripsByDate[tripDateString] = [];
            tripsByDate[tripDateString].push(trip);
          }
        });
      });
      return tripsByDate;
    }


    function useTripData(tripResults) {
      for (const channel in tripResults) {
        if (tripResults.hasOwnProperty(channel)) {
          let channelSeriesData = {name: vendor[channel], items: getEmptyItems()};
          const vendorData = tripResults[channel];
          const tripsByDate = filterByDate(vendorData);
          getSeriesData(vendorData, channelSeriesData);
          series.All.push(channelSeriesData);
          Object.keys(tripsByDate).forEach((dateKey) => {
            let channelSeriesDataForDate = {name: vendor[channel], items: getEmptyItems()};
            getSeriesDataForDate(tripsByDate[dateKey], channelSeriesDataForDate);
            if (series[dateKey]) {
              series[dateKey].push(channelSeriesDataForDate);
            } else {
              series[dateKey] = [];
              series[dateKey].push(channelSeriesDataForDate);
            }
          });
        }
      }
      // Assigning the value
      self.seriesValue(series.All);
    }

    self.stackValue = ko.observable('on');
    self.orientationValue = ko.observable('vertical');
    self.seriesValue = ko.observableArray();
    self.groupsValue = ko.observableArray(groups);

    self.changeDate = function() {
      let newKey = $('#tripsByDateOptions').val();
      self.seriesValue(series[newKey]);
    }
//============= End of area for line graph of trips by hour ====================


//============= Area for the AJAX request ======================================
      const requestOptions = {
        channels : [app.riderCoChannel, app.scooterCoChannel],
        url : app.cityRestProxy,
        chaincode : app.tripChaincode
      };

      const requestOptions2 = {
        channels : [app.riderCoChannel, app.scooterCoChannel],
        url : app.cityRestProxy,
        chaincode : app.vehicleChaincode
      };

      if (app.cloudChain) {
        requestOptions.restUsername = app.restUsername;
        requestOptions.restPassword = app.restPassword;
        requestOptions2.restUsername = app.restUsername;
        requestOptions2.restPassword = app.restPassword;
      };

      self.connected = function() {
        self.dateOptions = ko.observableArray(['All']);
        // AJAX for the trips.
        $.ajax({
          type: 'POST',
          url: 'http://localhost:3000/tripsCheck',
          data: JSON.stringify(requestOptions),
          contentType: "application/json",
          success: (data, textStatus, xhr) => {
            // Resetting the data so it doesn't get double-loaded.
            totalTrips = {
              All : 0
            };
            pieSeries = {
              All : [
                {name: "< 15 minutes", items: [0]},
                {name: "15 - 29 minutes", items: [0]},
                {name: "30 - 44 minutes", items: [0]},
                {name: "45 - 60 minutes", items: [0]},
                {name: "> 60 minutes", items: [0]}
              ]
            };
            series = {
              All : []
            };
            makePieChart(data);
            useTripData(data);
            self.dateOptions.sort();
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
        // AJAX for the vehicles
        $.ajax({
          type: 'POST',
          url: 'http://localhost:3000/vehiclesCheck',
          data: JSON.stringify(requestOptions2),
          contentType: "application/json",
          success: (data, textStatus, xhr) => {
            barSeries = [
              {
                name: "In-Service",
                items:[0, 0]
              },
              {
                name: "Out-of-Service",
                items:[0, 0]
              }
            ];
            useVehicleData(data);
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
    };
    self.transitionCompleted = function() {
    };
  }

  return new CityViewModel();
});

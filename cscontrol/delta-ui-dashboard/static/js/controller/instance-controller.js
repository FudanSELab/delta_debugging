var instance = angular.module('app.instance-controller', []);

instance.controller('InstanceCtrl', ['$scope', '$http','$window','loadTestCases','loadServiceList', 'serviceDeltaService',
        function($scope, $http,$window,loadTestCases,loadServiceList, serviceDeltaService) {

        // select testcases, remain the involved serves and stop other services
        // $scope.extractService = function(){
        //     var checkedTest = $("input[name='case']:checked");
        //     var tests = [];
        //     checkedTest.each(function(){
        //         tests.push($(this).val());
        //     });
        //     if(tests.length > 0){
        //         serviceDeltaService.deltaService(tests).then(function(result){
        //             console.log("============= service delta result ===============");
        //             console.log(result);
        //             console.log("==================================================");
        //         })
        //     }
        // };

        //刷新页面
        $scope.reloadRoute = function () {
            $window.location.reload();
        };

        // 加载service列表
        loadServiceList.loadServiceList().then(function (result) {
            if(result.status){
                $scope.services = result.services;
                $scope.serviceGroup = [];
                for(var i = 0; i < $scope.services.length; i++){
                    for(var j = 0; j < 5 && i < $scope.services.length; ){
                        if($scope.services[i].serviceName.indexOf("service") !== -1){
                            $scope.services[i].checked = false;
                            $scope.serviceGroup.push($scope.services[i]);
                            i++;
                            j++;
                        } else {
                            i++;
                        }
                    }
                }
            } else {
                alert(result.message);
            }
        });

        // $scope.test = function(){
        //     var checkedTest = $("input[name='testcase']:checked");
        //     var tests = [];
        //     checkedTest.each(function(){
        //         tests.push($(this).val());
        //     });
        //     console.log(tests);
        // };

        // 加载testcase列表
        loadTestCases.loadTestList().then(function (result) {
            $scope.testCases = result;
            // $scope.testCases = [];
            // for(var i = 0; i < result[0].products.length; i++){
            //     result[0].products[i].checked = false;
            //     $scope.testCases.push(result[0].products[i]);
            // }
        });

        var stompClient = null;
        //传递用户key值
        var loginId = new UUID().id;
        $scope.deltaResults = [];

        function setConnected(connected) {
            if(connected){
                $('#test-button').css('display', 'block');
            } else {
                $('#test-button').css('display', 'none');
            }
        }

        function connect() {
            var socket = new SockJS('/delta-socket');
            stompClient = Stomp.over(socket);
            stompClient.connect({login:loginId}, function (frame) {
                setConnected(true);
                // console.log('Connected: ' + frame);
                stompClient.subscribe('/user/topic/deltaresponse', function (data) {
                    // console.log("data.body--------\n");
                    // console.log(data.body);
                    var data = JSON.parse(data.body);
                    if(data.status){
                        var env = data.env;
                        var result = data.result.deltaResults;
                        var entry = {
                            services:"",
                            tests: "",
                            diff:false
                        } ;
                        if(data.diffFromFirst){
                            entry.diff = true;
                        }
                        for(var i = 0; i < env.length; i++){
                            entry.services += env[i].serviceName + ": " + env[i].numOfReplicas + "   ";
                        }
                        for(var j = 0; j < result.length; j++){
                            entry.tests += result[j].className + ": " + result[j].status + ";   " ;
                        }
                        // console.log("entry:\n");
                        // console.log(entry);
                        $scope.deltaResults.push(entry);
                        $scope.$apply();
                    } else {
                        alert(data.message);
                    }

                });
            });
        }

        function disconnect() {
            if ( stompClient != null ) {
                stompClient.disconnect();
            }
            setConnected(false);
            console.log("Disconnected");
        }

        $scope.sendDeltaData = function() {
            $scope.deltaResults = [];

            var checkedTest = $("input[name='testcase']:checked");
            var tests = [];
            checkedTest.each(function(){
                tests.push($(this).val());
            });

            var checkedServices = $("input[name='service']:checked");
            var env = [];
            checkedServices.each(function(){
                env.push($(this).val());
            });

            var data = {
                'id': loginId,
                'env': env,
                'tests': tests
            };
            console.log("data:\n");
            console.log(data);
            stompClient.send("/app/msg/delta", {}, JSON.stringify(data));
        }

        $scope.showDelta = function(){
            $scope.deltaResults = [];
            connect();
        };

        $window.onbeforeunload = function(){
            disconnect();
        };

            // $scope.test = function(){
            //     var checkedServices = $("input[name='service']:checked");
            //     var env = [];
            //     checkedServices.each(function(){
            //         env.push($(this).val());
            //     });
            //     console.log(env);
            // };

            // $scope.serviceGroup = [
            //     {
            //         "serviceName": "redis",
            //         "numOfReplicas": 1,
            //         "checked":false
            //     },
            //
            //     {
            //         "serviceName": "ts-route-service",
            //         "numOfReplicas": 1,
            //         "checked":false
            //     },
            //     {
            //         "serviceName": "ts-seat-service",
            //         "numOfReplicas": 1,
            //         "checked":false
            //     },
            //     {
            //         "serviceName": "ts-security-mongo",
            //         "numOfReplicas": 1,
            //         "checked":false
            //     },
            //     {
            //         "serviceName": "ts-security-service",
            //         "numOfReplicas": 1,
            //         "checked":false
            //     },
            //     {
            //         "serviceName": "ts-sso-service",
            //         "numOfReplicas": 1,
            //         "checked":false
            //     },
            //     {
            //         "serviceName": "ts-station-mongo",
            //         "numOfReplicas": 1,
            //         "checked":false
            //     },
            //     {
            //         "serviceName": "ts-station-service",
            //         "numOfReplicas": 1,
            //         "checked":false
            //     },
            //     {
            //         "serviceName": "ts-ticket-office-mongo",
            //         "numOfReplicas": 1,
            //         "checked":false
            //     },
            //     {
            //         "serviceName": "ts-ticket-office-service",
            //         "numOfReplicas": 1,
            //         "checked":false
            //     },
            //     {
            //         "serviceName": "ts-ticketinfo-service",
            //         "numOfReplicas": 1,
            //         "checked":false
            //     },
            //
            //     {
            //         "serviceName": "zipkin",
            //         "numOfReplicas": 1,
            //         "checked":false
            //     }
            // ];
            //
            // $scope.testNames=[
            //     {
            //         title:"23333"
            //     },
            //     {
            //         title:"2333"
            //     },
            //     {
            //         title:"233"
            //     }
            // ];
}]);

instance.factory('loadTestCases', function ($http, $q) {

    var service = {};

    //获取并返回数据
    service.loadTestList = function () {
        var deferred = $q.defer();
        var promise = deferred.promise;

        $http({
            method: "get",
            url: "/testBackend/getFileTree",
            contentType: "application/json",
            dataType: "json",
            withCredentials: true,
        }).success(function (data, status, headers, config) {
            if (data) {
                deferred.resolve(data);
            }
            else{
                alert("Request the order list fail!" + data.message);
            }
        });
        return promise;
    };
    return service;
});


instance.factory('loadServiceList', function ($http, $q) {
    var service = {};
    //获取并返回数据
    service.loadServiceList = function () {
        var deferred = $q.defer();
        var promise = deferred.promise;
        $http({
            method: "get",
            url: "/api/getServicesList",
            contentType: "application/json",
            dataType: "json",
            withCredentials: true
        }).success(function (data, status, headers, config) {
            if (data) {
                deferred.resolve(data);
            }
            else{
                alert("Request the order list fail!" + data.message);
            }
        });
        return promise;
    };
    return service;
});

instance.factory('serviceDeltaService', function ($http, $q) {

    var service = {};

    service.deltaService = function (testCaseList) {
        var deferred = $q.defer();
        var promise = deferred.promise;

        var checkedTest = $("input[name='case']:checked");
        var tests = [];
        checkedTest.each(function(){
            tests.push($(this).val());
        });

        $http({
            method: "post",
            url: "/delta/extractServices",
            contentType: "application/json",
            dataType: "json",
            data: {
                tests:testCaseList
            },
            withCredentials: true
        }).success(function (data) {
            if (data) {
                deferred.resolve(data);
            }
            else{
                alert("Request the order list fail!" + data.message);
            }
        });
        return promise;
    };
    return service;
});
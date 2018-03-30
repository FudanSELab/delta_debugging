var service = angular.module('app.service-controller', []);

service.controller('ServiceCtrl',['$scope', '$http','$window','loadTestCases',  'serviceLogService', 'refreshPodsService','getPodLogService',
    function($scope, $http,$window,loadTestCases,  serviceLogService, refreshPodsService, getPodLogService) {

    // 加载testcase列表
    loadTestCases.loadTestList().then(function (result) {
        $scope.testCases = result;
    });

    refreshPodsService.load().then(function(result){
       if(result.status){
            $scope.podList = result.pods;
       } else {
           alert(result.message);
       }
    });

    $scope.refreshPod = function(){
        refreshPodsService.load().then(function(result){
            if(result.status){
                $scope.podList = result.pods;
            } else {
                alert(result.message);
            }
        });
    };

    $scope.reservedServices = "delta result...";
    var stompClient = null;
    //传递用户key值
    var loginId = new UUID().id;

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
            stompClient.subscribe('/user/topic/serviceDeltaResponse', function (data) {
                $('#test-button').removeClass('disabled');
                var data = JSON.parse(data.body);
                if(data.status){
                    $scope.reservedServices = JSON.stringify(data.serviceNames);
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

    $scope.extractService = function() {
        $scope.reservedServices = "testing...";
        var checkedTest = $("input[name='testcase']:checked");
        var tests = [];
        checkedTest.each(function(){
            tests.push($(this).val());
        });
        if(tests.length > 0){
            $('#test-button').addClass('disabled');
            var data = {
                'id': loginId,
                'tests': tests
            };
            console.log("data:\n");
            console.log(data);
            stompClient.send("/app/msg/serviceDelta", {}, JSON.stringify(data));
        } else {
            alert("To delete node, please select at least one testcase.");
        }
    };

    $scope.showDelta = function(){
        $scope.reservedServices = "";
        connect();
    };

    $window.onbeforeunload = function(){
        disconnect();
    };

    $scope.servicelogs = "";
    $scope.getPodLogs = function(){
        if ( stompClient != null ) {
            var checkedPods = $("input[name='pod']:checked");
            var pods = [];
            checkedPods.each(function () {
                pods.push($(this).val());
            });
            if(pods.length > 0){
                $('#inspectPodButton').addClass('disabled');
                getPodLogService.load(pods[0]).then(function(result){
                    if(result.status){
                        $scope.servicelogs += result.podLog.podName +  ":</br>" + result.podLog.logs + "</br>";
                        var height = $('#service-logs').prop('scrollHeight');
                        $('#service-logs').scrollTop(height);
                        $('#inspectPodButton').removeClass('disabled');
                    } else {
                        alert(result.message);
                    }
                })
            } else {
                alert("Please check at least one pod to show the logs!");
            }
        } else {
            alert("Please click the connect button.")
        }


    };


}]);


service.factory('serviceLogService', function ($http, $q) {
    var service = {};
    service.loadLogs = function () {
        var deferred = $q.defer();
        var promise = deferred.promise;
        // $http({
        //     method: "post",
        //     url: "/xxx/xxx",
        //     contentType: "application/json",
        //     dataType: "json",
        //     withCredentials: true
        // }).success(function (data) {
        //     if (data) {
        //         deferred.resolve(data);
        //     } else{
        //         alert("Get logs fail!" + data.message);
        //     }
        // });
        deferred.resolve("2333");
        return promise;
    };
    return service;
});


// service.factory('serviceDeltaService', function ($http, $q) {
//     var service = {};
//     service.deltaService = function (testCaseList) {
//         var deferred = $q.defer();
//         var promise = deferred.promise;
//
//         $http({
//             method: "post",
//             url: "/delta/extractServices",
//             contentType: "application/json",
//             dataType: "json",
//             data: {
//                 tests:testCaseList
//             },
//             withCredentials: true
//         }).success(function (data) {
//             if (data) {
//                 deferred.resolve(data);
//             }
//             else{
//                 alert("Request the order list fail!" + data.message);
//             }
//         });
//         return promise;
//     };
//     return service;
// });



// service.factory('loadTestCases', function ($http, $q) {
//     var service = {};
//     //获取并返回数据
//     service.loadTestList = function () {
//         var deferred = $q.defer();
//         var promise = deferred.promise;
//
//         $http({
//             method: "get",
//             url: "/testBackend/getFileTree",
//             contentType: "application/json",
//             dataType: "json",
//             withCredentials: true,
//         }).success(function (data, status, headers, config) {
//             if (data) {
//                 deferred.resolve(data);
//             }
//             else{
//                 alert("Request the order list fail!" + data.message);
//             }
//         });
//         return promise;
//     };
//     return service;
// });
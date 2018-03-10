var service = angular.module('app.service-controller', []);

service.controller('ServiceCtrl',['$scope', '$http','$window','loadTestCases',
    function($scope, $http,$window,loadTestCases) {

    // 加载testcase列表
    loadTestCases.loadTestList().then(function (result) {
        $scope.testCases = [];
        for(var i = 0; i < result[0].products.length; i++){
            result[0].products[i].checked = false;
            $scope.testCases.push(result[0].products[i]);
        }
    });

        $scope.reservedServices ="";
        // $scope.extractService = function () {
        //     var checkedTest = $("input[name='testcase']:checked");
        //     var tests = [];
        //     checkedTest.each(function () {
        //         tests.push($(this).val());
        //     });
        //     if (tests.length > 0) {
        //         serviceDeltaService.deltaService(tests).then(function (result) {
        //             console.log("============= service delta result ===============");
        //             console.log(result);
        //             console.log("==================================================");
        //             if(result.status){
        //                 $scope.reservedServices = JSON.stringify(result.serviceNames);
        //             } else {
        //                 alert(result.message);
        //             }
        //         })
        //     }
        // };


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
            stompClient.subscribe('/user/topic/serviceDeltaResponse', function (data) {
                // console.log("data.body--------\n");
                // console.log(data.body);
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
        var data = {
            'id': loginId,
            'tests': tests
        };
        console.log("data:\n");
        console.log(data);
        stompClient.send("/app/msg/serviceDelta", {}, JSON.stringify(data));
    };

    $scope.showDelta = function(){
        $scope.reservedServices = "";
        connect();
    };


    $window.onbeforeunload = function(){
        disconnect();
    };


}]);


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



service.factory('loadTestCases', function ($http, $q) {
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
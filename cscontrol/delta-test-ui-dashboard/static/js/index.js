var app = angular.module("myApp",[]);

app.factory('loadDataService', function ($http, $q) {

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


app.factory('loadDataService2', function ($http, $q) {
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


app.controller('indexCtrl', function ($scope, $http,$window,loadDataService,loadDataService2) {

    //刷新页面
    $scope.reloadRoute = function () {
        $window.location.reload();
    };

    var deltaItems = [
        {
            title: 'Types',
            type: 'folder',
            products: [
                {
                    title: 'Instance',
                    type: 'item'
                },
                {
                    title: 'Node',
                    type: 'item'
                },
                {
                    title: 'Config',
                    type: 'item'
                }
            ]
        }
    ];

    // $scope.serviceNames = ['s1', 's2', 's3','s4', 's5', 's6', 's7', 's8','s9', 's10'];

    // 加载service列表
    loadDataService2.loadServiceList().then(function (result) {
        // console.log("services--------\n");
        // console.log(result);
        if(result.status){
            $scope.serviceNames = result.services;
            $scope.serviceNamesGroup = [];
            for(var i = 0; i < $scope.serviceNames.length; i++){
                var temp = [];
                for(var j = 0; j < 5; j++){
                    if(i<$scope.serviceNames.length){
                        temp.push($scope.serviceNames[i]);
                        i++
                    } else {
                        break;
                    }
                }
                $scope.serviceNamesGroup.push(temp);
            }
            //$scope.$apply();
        } else {
            alert(result.message);
        }
    });



    // $scope.test=function(){
    //     var checked = $("input[name='service']:checked");
    //     var tests = [];
    //     checked.each(function(){
    //         tests.push($(this).val());
    //     });
    //     console.log(tests);
    // };


    $('#fileTree').tree({
        dataSource:function(options, callback) {
            // 模拟异步加载
            //options.products一定要！不然会无限循环套用
            callback({data: options.products || deltaItems});
        },
        multiSelect: false,
        cacheItems: true,
        folderSelect: false
    });


    // 加载testcase列表
    loadDataService.loadTestList().then(function (result) {
        // console.log(JSON.stringify(result));
        $scope.testNames = result[0].products;
    });

    $('#fileTree').on('selected.tree.amui', function (event, data) {
        // console.log(data);
        $scope.testName = data.target.title;
        $scope.resultCount = "";
        $scope.results = [];
    });

    var stompClient = null;
    //传递用户key值
    var loginId = new UUID().id;
    $scope.deltaResults = [];

    function setConnected(connected) {
        $('#test-button').css('display', 'block');
    }

    function connect() {
        var socket = new SockJS('/delta-socket');
        stompClient = Stomp.over(socket);
        stompClient.connect({login:loginId}, function (frame) {
            setConnected(true);
            console.log('Connected: ' + frame);
            stompClient.subscribe('/user/topic/deltaresponse', function (data) {
                console.log("data.body--------\n");
                console.log(data.body);
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
                    console.log("entry:\n");
                    console.log(entry);
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

    function sendDeltaData() {
        $scope.deltaResults = [];

        var checkedTest = $("input[name='case']:checked");
        var tests = [];
        checkedTest.each(function(){
            tests.push($(this).val());
        });

        var checkedServices = $("input[name='service']:checked");
        var env = [];
        checkedServices.each(function(){
            env.push($(this).val());
        });

        // var data = {
        //     'id': loginId,
        //     'env': [
        //         {'serviceName': 's1', 'instanceNum': 3},
        //         {'serviceName': 's2', 'instanceNum': 3}
        //     ],
        //     'tests': tests
        // };
        var data = {
            'id': loginId,
            'env': env,
            'tests': tests
        };
        console.log("data:\n");
        console.log(data);
        stompClient.send("/app/msg/deltatest", {}, JSON.stringify(data));
    }


    $scope.showDelta = function(){
        if( "Instance" == $scope.testName){
            $scope.deltaResults = [];
            connect();
        }
    };

    $scope.startDeltaTest = function(){
        sendDeltaData();
    };

    $window.onbeforeunload = function(){
        disconnect();
    };

    // $scope.$on('$destroy', function() {
    //     $window.onbeforeunload = undefined;
    // });


});
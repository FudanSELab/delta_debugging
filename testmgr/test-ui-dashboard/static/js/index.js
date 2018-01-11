var app = angular.module("myApp",[]);

app.factory('loadDataService', function ($http, $q) {

    var service = {};

    //获取并返回数据
    service.loadRecordList = function () {
        var deferred = $q.defer();
        var promise = deferred.promise;

        $http({
            method: "get",
            url: "/testBackend/getFileTree",
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

app.controller('indexCtrl', function ($scope, $http,$window,loadDataService) {

    //刷新页面
    $scope.reloadRoute = function () {
        $window.location.reload();
    };

    //首次加载显示数据
    loadDataService.loadRecordList().then(function (result) {
        // console.log(JSON.stringify(result));
        $('#fileTree').tree({
            dataSource:function(options, callback) {
                // 模拟异步加载
                //options.products一定要！不然会无限循环套用
                callback({data: options.products || result});
            },
            multiSelect: false,
            cacheItems: true,
            folderSelect: false
        });
    });

    $('#fileTree').on('selected.tree.amui', function (event, data) {
        // console.log(data);
        $scope.testName = data.target.title.split(".")[0];
        $scope.resultCount = "";
        $scope.results = [];
    });

    $scope.startTest = function(){
        if( null != $scope.testName && "" != $scope.testName){
            $http({
                method: "post",
                url: "/testBackend/test",
                data:{testString:$scope.testName},
                withCredentials: true
            }).success(function (data, status, headers, config) {
                console.log(data);
                if(data.status){
                    $scope.results = data.resultList;
                    var count = data.resultCount;
                    var all = parseInt(count[0]) + parseInt(count[1]) + parseInt(count[2]) + parseInt(count[3]);
                    $scope.resultCount = count[0] + "/" + all + "   " + data.message;
                } else {
                    $scope.resultCount = data.message;
                }
            });
        }
    }


    //  /msg/sendcommuser
    var stompClient = null;
    //传递用户key值
    var login = new UUID();

    function setConnected(connected) {
        $('#test-button').css('display', 'block');
    }

    function connect() {
        var socket = new SockJS('/testBackend/ricky-websocket');
        stompClient = Stomp.over(socket);
        stompClient.connect({login:login}, function (frame) {
            setConnected(true);
            console.log('Connected: ' + frame);
            stompClient.subscribe('/user/topic/greetings', function (greeting) {
                // console.log("greeting:\n");
                // console.log(greeting);
                $scope.deltaResults = $scope.deltaResults.concat(JSON.parse(greeting.body).testResult.resultList);
                console.log($scope.deltaResults);
                $scope.$apply();
                // showGreeting(JSON.parse(greeting.body));
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

    function sendName() {
        stompClient.send("/app/msg/hellosingle", {}, JSON.stringify(login));
    }

    function showGreeting(message) {
        // alert(message);
        console.log(message);
        // $("#greetings").append("<tr><td>" + message + "</td></tr>");
    }


    $scope.showDelta = function(){
        if( null != $scope.testName && "" != $scope.testName){
            $scope.deltaResults = [];
            connect();
        }
    };

    $scope.startDeltaTest = function(){
        sendName();
    };

    $scope.beforeunload = function(){
        disconnect();
    };


});
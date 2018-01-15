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

    // $scope.startTest = function(){
    //     if( null != $scope.testName && "" != $scope.testName){
    //         $http({
    //             method: "post",
    //             url: "/testBackend/test",
    //             data:{testString:$scope.testName},
    //             withCredentials: true
    //         }).success(function (data, status, headers, config) {
    //             console.log(data);
    //             if(data.status){
    //                 $scope.results = data.resultList;
    //                 var count = data.resultCount;
    //                 var all = parseInt(count[0]) + parseInt(count[1]) + parseInt(count[2]) + parseInt(count[3]);
    //                 $scope.resultCount = count[0] + "/" + all + "   " + data.message;
    //             } else {
    //                 $scope.resultCount = data.message;
    //             }
    //         });
    //     }
    // }


    //  /msg/sendcommuser
    var stompClient = null;
    //传递用户key值
    var loginId = new UUID().id;

    function setConnected(connected) {
        $('#test-button').css('display', 'block');
    }

    function connect() {
        var socket = new SockJS('/test-socket');
        stompClient = Stomp.over(socket);
        stompClient.connect({login:loginId}, function (frame) {
            setConnected(true);
            console.log('Connected: ' + frame);

            stompClient.subscribe('/user/topic/testresponse', function (data) {
                console.log(data.body);
                var response = JSON.parse(data.body);
                if(response.status){
                    $scope.testResults = response.testResult.resultList;
                    var count = response.testResult.resultCount;
                    var all = parseInt(count[0]) + parseInt(count[1]) + parseInt(count[2]) + parseInt(count[3]);
                    $scope.resultCount = count[0] + "/" + all + "   " + response.message;
                } else {
                    $scope.resultCount = response.message;
                }
                $scope.$apply();
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


    $scope.showTestButton = function(){
        if( null != $scope.testName && "" != $scope.testName){
            $scope.testResults = [];
            connect();
        }
    };

    $scope.startTest = function(){
        var data = {
            id: loginId,
            testName: $scope.testName
        };
        stompClient.send("/app/msg/testsingle", {}, JSON.stringify(data));
    };

    $window.onbeforeunload = function(){
        disconnect();
    };



    // $scope.$on('$destroy', function() {
    //     $window.onbeforeunload = undefined;
    // });

});
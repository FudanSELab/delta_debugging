var node = angular.module('app.node-controller', []);

node.controller('NodeCtrl', ['$scope', '$http','$window','loadNodeList', 'refreshPodsService','getPodLogService',
    function($scope, $http,$window,loadNodeList, refreshPodsService, getPodLogService) {

    $scope.refreshNodeList = function(){
        // 加载node列表
        loadNodeList.load().then(function (result) {
            $scope.nodeList = result.nodes;
        });


        // $scope.nodeList = [
        //     {
        //         "role": "Master",
        //         "name": "centos-master",
        //         "ip": "10.141.211.181",
        //         "status": "Ready",
        //         "kubeProxyVersion": "v1.9.3",
        //         "kubeletVersion": "v1.9.3",
        //         "operatingSystem": "linux",
        //         "osImage": "CentOS Linux 7 (Core)",
        //         "containerRuntimeVersion": "docker://17.3.2"
        //     },
        //     {
        //         "role": "Minion",
        //         "name": "centos-minion-1",
        //         "ip": "10.141.211.179",
        //         "status": "Ready",
        //         "kubeProxyVersion": "v1.9.3",
        //         "kubeletVersion": "v1.9.3",
        //         "operatingSystem": "linux",
        //         "osImage": "CentOS Linux 7 (Core)",
        //         "containerRuntimeVersion": "docker://17.3.2"
        //     },
        //     {
        //         "role": "Minion",
        //         "name": "centos-minion-2",
        //         "ip": "10.141.211.180",
        //         "status": "Ready",
        //         "kubeProxyVersion": "v1.9.3",
        //         "kubeletVersion": "v1.9.3",
        //         "operatingSystem": "linux",
        //         "osImage": "CentOS Linux 7 (Core)",
        //         "containerRuntimeVersion": "docker://17.3.2"
        //     },
        //     {
        //         "role": "Minion",
        //         "name": "centos-minion-3",
        //         "ip": "10.141.211.173",
        //         "status": "Ready",
        //         "kubeProxyVersion": "v1.9.3",
        //         "kubeletVersion": "v1.9.3",
        //         "operatingSystem": "linux",
        //         "osImage": "CentOS Linux 7 (Core)",
        //         "containerRuntimeVersion": "docker://17.3.2"
        //     }
        // ];

    };

    $scope.refreshNodeList();


    /*resfresh pod list*/
    refreshPodsService.load().then(function(result){
        if(result.status){
            $scope.podList = result.pods;
        } else {
            alert(result.message);
        }
    });

    $scope.refreshPod = function(){
        $('#refreshPodButton').addClass('disabled');
        refreshPodsService.load().then(function(result){
            // alert("23333");
            if(result.status){
                $scope.podList = result.pods;
            } else {
                alert(result.message);
            }
            $('#refreshPodButton').removeClass('disabled');
        });
    };

    $scope.deleteResult = "delta result...";

    var stompClient = null;
    //传递用户key值
    var loginId = new UUID().id;
    // $scope.deleteResult = [];

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
            stompClient.subscribe('/user/topic/nodeDeltaResponse', function (data) {
                $('#test-button').removeClass('disabled');
                var data = JSON.parse(data.body);
                if(data.status){
                    $scope.deleteResult = JSON.stringify(data.message);
                    $scope.refreshNodeList();
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

    $scope.deleteNodes = function() {
        $scope.deleteResult = "testing...";
        var checkedNodes = $("input[name='node']:checked");
        var nodes = [];
        checkedNodes.each(function () {
            nodes.push($(this).val());
        });
        if(nodes.length > 0){
            $('#test-button').addClass('disabled');
            var data = {
                'id': loginId,
                'nodeNames': nodes
            };
            console.log("data:\n");
            console.log(data);
            stompClient.send("/app/msg/nodeDelta", {}, JSON.stringify(data));
        } else {
            alert("To delete node, please select at least one node.");
        }
    };

    $scope.showDelta = function(){
        $scope.deleteResult = "";
        connect();
    };

    $window.onbeforeunload = function(){
        disconnect();
    };


    $scope.nodelogs = "";
    $scope.getPodLogs = function(){
        var checkedPods = $("input[name='pod']:checked");
        var pods = [];
        checkedPods.each(function () {
            pods.push($(this).val());
        });
        if(pods.length > 0){
            $('#inspectPodButton').addClass('disabled');
            getPodLogService.load(pods[0]).then(function(result){
                if(result.status){
                    $scope.nodelogs += result.podLog.podName +  ":</br>" + result.podLog.logs + "</br>";
                    var height = $('#node-logs').prop('scrollHeight');
                    $('#node-logs').scrollTop(height);
                    $('#inspectPodButton').removeClass('disabled');
                } else {
                    alert(result.message);
                }
            })
        } else {
            alert("Please check at least one pod to show the logs!");
        }
    };


}]);

// node.factory('nodeLogService', function ($http, $q) {
//     var service = {};
//     service.loadLogs = function () {
//         var deferred = $q.defer();
//         var promise = deferred.promise;
//         // $http({
//         //     method: "post",
//         //     url: "/xxx/xxx",
//         //     contentType: "application/json",
//         //     dataType: "json",
//         //     withCredentials: true
//         // }).success(function (data) {
//         //     if (data) {
//         //         deferred.resolve(data);
//         //     } else{
//         //         alert("Get logs fail!" + data.message);
//         //     }
//         // });
//         deferred.resolve("2333");
//         return promise;
//     };
//     return service;
// });


// node.factory('loadNodeList', function ($http, $q) {
//     var service = {};
//     //获取并返回数据
//     service.load = function () {
//         var deferred = $q.defer();
//         var promise = deferred.promise;
//
//         $http({
//             method: "get",
//             url: "/api/getNodesList",
//             contentType: "application/json",
//             dataType: "json",
//             withCredentials: true
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



// node.factory('nodeDeltaService', function ($http, $q) {
//     var service = {};
//     //获取并返回数据
//     service.delta = function (nodes) {
//         var deferred = $q.defer();
//         var promise = deferred.promise;
//         $http({
//             method: "post",
//             url: "/delta/deleteNodes",
//             contentType: "application/json",
//             dataType: "json",
//             data:nodes,
//             withCredentials: true,
//         }).success(function (data, status, headers, config) {
//             if (data) {
//                 deferred.resolve(data);
//             }
//             else{
//                 alert("Delete the node fail!" + data.message);
//             }
//         });
//         return promise;
//     };
//     return service;
// });
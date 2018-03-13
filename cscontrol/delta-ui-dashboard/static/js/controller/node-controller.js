var node = angular.module('app.node-controller', [])

node.controller('NodeCtrl', ['$scope', '$http','$window','loadNodeList',
    function($scope, $http,$window,loadNodeList) {

    $scope.refreshNodeList = function(){
        // 加载node列表
        loadNodeList.load().then(function (result) {
            $scope.nodeList = [];
            for(var i = 0; i < result.nodes.length; i++){
                result.nodes[i].checked = false;
                $scope.nodeList.push(result.nodes[i]);
            }
        });

        // console.log("23333");
        // $scope.nodeList = [
        //     {
        //         "role": "Minion",
        //         "name": "centos-minion-1",
        //         "ip": "10.141.211.179",
        //         "status": "Ready",
        //         "kubeProxyVersion": "v1.9.3",
        //         "kubeletVersion": "v1.9.3",
        //         "operatingSystem": "linux",
        //         "osImage": "CentOS Linux 7 (Core)",
        //         "containerRuntimeVersion": "docker://17.3.2",
        //         "checked":false
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
        //         "containerRuntimeVersion": "docker://17.3.2",
        //         "checked":false
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
        //         "containerRuntimeVersion": "docker://17.3.2",
        //         "checked":false
        //     }
        // ];
    };

    $scope.refreshNodeList();

    $scope.deleteResult = "delta result...";

    // $scope.deleteNodes = function () {
    //     var checkedNodes = $("input[name='node']:checked");
    //     var nodes = [];
    //     checkedNodes.each(function () {
    //         nodes.push($(this).val());
    //     });
    //     // console.log(nodes);
    //     if (nodes.length > 0) {
    //         nodeDeltaService.delta(nodes).then(function (result) {
    //             console.log("============= service delta result ===============");
    //             console.log(result);
    //             console.log("==================================================");
    //             if(result.status){
    //                 $scope.deleteResult = JSON.stringify(result.messages);
    //                 $scope.refreshNodeList();
    //             } else {
    //                 alert(result.message);
    //             }
    //         })
    //     }
    // };

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
            // console.log('Connected: ' + frame);
            stompClient.subscribe('/user/topic/nodeDeltaResponse', function (data) {
                // console.log("data.body--------\n");
                // console.log(data.body);
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
            var data = {
                'id': loginId,
                'nodeNames': nodes
            };
            console.log("data:\n");
            console.log(data);
            stompClient.send("/app/msg/nodeDelta", {}, JSON.stringify(data));
        }
    };

    $scope.showDelta = function(){
        $scope.deleteResult = "";
        connect();
    };

    $window.onbeforeunload = function(){
        disconnect();
    };


}]);


node.factory('loadNodeList', function ($http, $q) {
    var service = {};
    //获取并返回数据
    service.load = function () {
        var deferred = $q.defer();
        var promise = deferred.promise;

        $http({
            method: "get",
            url: "/api/getNodesList",
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
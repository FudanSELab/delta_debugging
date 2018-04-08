var config = angular.module('app.config-controller', []);

config.controller('ConfigCtrl', ['$scope', '$http','$window','loadServiceList',  'getConfigService','getPodLogService', 'refreshPodsService',
    function($scope, $http,$window,loadServiceList,  getConfigService, getPodLogService, refreshPodsService) {

        $scope.test = function(){
            var checkedConfig = $("input[name='config']:checked");
            var configs = [];
            checkedConfig.each(function(){
                configs.push($(this).val());
            });

            console.log("configs:\n" );
            console.log(configs);
        };

        //刷新页面
        $scope.reloadRoute = function () {
            $window.location.reload();
        };

        // 加载config
        getConfigService.load().then(function (result) {
            if(result.status){
                // $scope.clusterConfig = result.data.clusterConfig;
                $scope.serviceConfig = result.services;
            } else {
                alert(result.message);
            }
        });

        // $scope.test = function(){
        //     var checkedConfig = $("input[name='config']:checked");
        //     var configs = [];
        //     checkedConfig.each(function(){
        //         configs.push($(this).val());
        //     });
        //
        //     console.log("configs:\n" );
        //     console.log(configs);
        // };


        //load pods
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


        var stompClient = null;
        //传递用户key值
        var loginId = new UUID().id;
        $scope.configDeltaResult = "config delta test...";

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
                stompClient.subscribe('/user/topic/configDeltaResponse', function (data) {
                    var data = JSON.parse(data.body);
                    if(data.status){
                        $scope.configDeltaResult = JSON.stringify(data.message);
                        //todo:refresh
                        $scope.$apply();
                    } else {
                        alert(data.message);
                    }
                });

            });
        }


        $scope.sendDeltaData = function() {
            $scope.configDeltaResult = "";
            var checkedConfig = $("input[name='config']:checked");
            var configs = [];
            checkedConfig.each(function(){
                configs.push($(this).val());
            });

            console.log("configs:\n" );
            console.log(configs);



            if(configs.length > 0){
                $('#test-button').addClass('disabled');
                var data = {
                    'id': loginId,
                   'configs': configs
                };
                stompClient.send("/app/msg/configDelta", {}, JSON.stringify(data));
            } else {
                alert("config delta failed!");
            }

        };

        $scope.showDelta = function(){
            $scope.configDeltaResults = "";
            connect();
        };

        function disconnect() {
            if ( stompClient != null ) {
                stompClient.disconnect();
            }
            setConnected(false);
            console.log("Disconnected");
        }

        $window.onbeforeunload = function(){
            disconnect();
        };

        $scope.configlogs = "";
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
                            $scope.configlogs += result.podLog.podName +  ":</br>" + result.podLog.logs + "</br>";
                            var height = $('#config-logs').prop('scrollHeight');
                            $('#config-logs').scrollTop(height);
                            $('#inspectPodButton').removeClass('disabled');
                        } else {
                            alert(result.message);
                        }
                    });
                } else {
                    alert("Please check at least one pod to show its logs!");
                }
            } else {
                alert("Please click the connect button.")
            }

        };


    }]);






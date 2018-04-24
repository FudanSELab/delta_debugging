var config = angular.module('app.config-controller', []);

config.controller('ConfigCtrl', ['$scope', '$http','$window','loadServiceList',  'getConfigService','getPodLogService', 'refreshPodsService','loadTestCases',
    function($scope, $http,$window,loadServiceList,  getConfigService, getPodLogService, refreshPodsService, loadTestCases) {

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
        $scope.refreshConfigs = function(){
            getConfigService.load().then(function (result) {
                if(result.status){
                    // $scope.clusterConfig = result.data.clusterConfig;
                    $scope.serviceConfig = result.services;
                } else {
                    alert(result.message);
                }
            });
        };
        $scope.refreshConfigs();

        // 加载testcase列表
        loadTestCases.loadTestList().then(function (result) {
            $scope.testCases = result;
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
        $scope.configDeltaResponse = [];

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
                    console.log("\n get response:");
                    console.log(data);
                    if(data.status){
                        var env = data.env;
                        var result = data.result.deltaResults;
                        var entry = {
                            configs:"",
                            tests: ""
                        } ;
                        for(var i = 0; i < env.length; i++){
                            entry.configs += env[i].serviceName + ": " + env[i].type + ":{" + env[i].key + ": " + env[i].value + "};     ";
                        }
                        for(var j = 0; j < result.length; j++){
                            entry.tests += result[j].className + ": " + result[j].status + ";   " ;
                        }
                        $scope.configDeltaResponse.push(entry);
                        $scope.$apply();
                    } else {
                        console.log("configDeltaResponse" + data.message);
                    }
                });

                stompClient.subscribe('/user/topic/configDeltaEnd', function (data) {
                    var data = JSON.parse(data.body);
                    console.log("\n end:");
                    console.log(data);
                    if(data.status){
                        alert("ddmingResult: " + data.ddminResult );
                        $scope.configDeltaResult = data.ddminResult;
                        // console.log("data.ddminResult: " + data.ddminResult);
                        $scope.$apply();
                    } else {
                        alert(data.message);
                    }
                    $('#test-button').removeClass('disabled');
                });


                stompClient.subscribe('/user/topic/simpleSetOrignalResult', function (data) {
                    alert(data.body);
                    $scope.refreshConfigs();
                    $('#setOriginal').removeClass('disabled');
                });

            });
        }


        $scope.sendDeltaData = function() {
            var checkedConfig = $("input[name='serviceconfig']:checked");
            var configs = [];
            checkedConfig.each(function(){
                var temp = $(this).val().split(":");
                configs.push({
                    'serviceName': temp[0],
                    'type': temp[1],
                    'key': temp[2],
                    'value': temp[3]
                });
            });
            // var finalConfigs = [];
            // configs.forEach(function(c){
            //     var existService;
            //     finalConfigs.forEach(function(f){
            //         if(f.serviceName == c.serviceName) existService = f;
            //     });
            //    if(existService) {
            //        var hasType = 0;
            //        existService['configs'].forEach(function(t){
            //            if(t.type == c.type){
            //                t.values.push({
            //                    'key': c.key,
            //                    'value': c.value
            //                });
            //                hasType = 1;
            //            }
            //        });
            //        if(hasType == 0){
            //            existService['configs'].push({
            //                'type': c.type,
            //                'values': [{
            //                    'key': c.key,
            //                    'value': c.value
            //                }]
            //            })
            //        }
            //    } else {
            //        finalConfigs.push({
            //            'serviceName':c.serviceName,
            //            'configs': [{
            //                 'type': c.type,
            //                 'values': [{
            //                    'key': c.key,
            //                    'value': c.value
            //                 }]
            //            }]
            //        })
            //    }
            // });
            // console.log("finalConfigs:");
            // console.log(JSON.stringify(finalConfigs));

            var checkedTest = $("input[name='testcase']:checked");
            var tests = [];
            checkedTest.each(function(){
                tests.push($(this).val());
            });

            if(configs.length > 0 && tests.length > 0 ){
                $('#test-button').addClass('disabled');
                $scope.configDeltaResponse = [];
                $scope.configDeltaResult = "config testing...";
                var data = {
                    'id': loginId,
                    'configs': configs,
                    'tests': tests
                };
                stompClient.send("/app/msg/configDelta", {}, JSON.stringify(data));
            } else {
                alert("Please select at least one config item and one test case!");
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

        $scope.simpleSetOrignal = function(n){
            if ( stompClient != null ) {
                var checkedConfig = $("input[name='serviceconfig']:checked");
                var configs = [];
                checkedConfig.each(function(){
                    var temp = $(this).val().split(":");
                    var v;
                    if(temp[2] == 'memory'){
                        v = "350Mi";
                    } else if(temp[2] == 'cpu'){
                        v = "300m";
                    } else {
                        alert("the key cannot be mapped to a default value.");
                    }
                    configs.push({
                        'serviceName': temp[0],
                        'type': temp[1],
                        'key': temp[2],
                        'value': v
                    });
                });
                console.log("simpleSetOrignal configs: ");
                console.log(configs);
                // var finalConfigs = [];
                // configs.forEach(function(c){
                //     var existService;
                //     finalConfigs.forEach(function(f){
                //         if(f.serviceName == c.serviceName) existService = f;
                //     });
                //     if(existService) {
                //         var hasType = 0;
                //         existService['configs'].forEach(function(t){
                //             if(t.type == c.type){
                //                 t.values.push({
                //                     'key': c.key,
                //                     'value': c.value
                //                 });
                //                 hasType = 1;
                //             }
                //         });
                //         if(hasType == 0){
                //             existService['configs'].push({
                //                 'type': c.type,
                //                 'values': [{
                //                     'key': c.key,
                //                     'value': c.value
                //                 }]
                //             })
                //         }
                //     } else {
                //         finalConfigs.push({
                //             'serviceName':c.serviceName,
                //             'configs': [{
                //                 'type': c.type,
                //                 'values': [{
                //                     'key': c.key,
                //                     'value': c.value
                //                 }]
                //             }]
                //         })
                //     }
                // });
                // console.log("simpleSetOrignal finalConfigs: ");
                // console.log(finalConfigs);

                if(configs.length > 0){
                    $('#setOriginal').addClass('disabled');
                    var data = {
                        'id': loginId,
                        'configs':configs
                    };
                    stompClient.send("/app/msg/simpleSetOrignal", {}, JSON.stringify(data));
                } else {
                    alert("Please select at least one config.");
                }
            } else {
                alert("Please click the connect button.")
            }
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
                        } else {
                            alert(result.message);
                        }
                        $('#inspectPodButton').removeClass('disabled');
                    });
                } else {
                    alert("Please check at least one pod to show its logs!");
                }
            } else {
                alert("Please click the connect button.")
            }

        };


    }]);






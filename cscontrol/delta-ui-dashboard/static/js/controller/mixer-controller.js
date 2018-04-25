var mixer = angular.module('app.mixer-controller', []);

mixer.controller('MixerCtrl', ['$scope', '$http','$window','loadServiceList',  'getConfigService','getPodLogService', 'refreshPodsService','loadTestCases',
    function($scope, $http,$window,loadServiceList,  getConfigService, getPodLogService, refreshPodsService, loadTestCases) {

        $scope.refreshServices = function(){
            loadServiceList.loadServiceList().then(function (result) {
                if(result.status){
                    $scope.services = result.services;
                    $scope.serviceGroup = [];
                    $scope.senderGroup = [];
                    $scope.receiverGroup = [];
                    for(var i = 0; i < $scope.services.length; i++){
                        if($scope.services[i].serviceName.indexOf("service") !== -1){
                            $scope.serviceGroup.push($scope.services[i]);
                            $scope.senderGroup.push($scope.services[i]);
                            $scope.receiverGroup.push($scope.services[i]);
                        }
                    }
                } else {
                    alert(result.message);
                }
            });
        };
        // 加载instance service列表
        $scope.refreshServices();

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

        // //load pods
        // refreshPodsService.load().then(function(result){
        //     if(result.status){
        //         $scope.podList = result.pods;
        //     } else {
        //         alert(result.message);
        //     }
        // });
        //
        // $scope.refreshPod = function(){
        //     refreshPodsService.load().then(function(result){
        //         if(result.status){
        //             $scope.podList = result.pods;
        //         } else {
        //             alert(result.message);
        //         }
        //     });
        // };


        var stompClient = null;
        //传递用户key值
        var loginId = new UUID().id;
        $scope.mixerDeltaResult = "mixer delta testing...";
        $scope.mixerDeltaResponse = [];

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
                stompClient.subscribe('/user/topic/mixerDeltaResponse', function (data) {
                    var data = JSON.parse(data.body);
                    console.log("\n get response:");
                    console.log(data);
                    if(data.status){
                        var result = data.result.deltaResults;
                        var entry = {
                            env:"",
                            tests: ""
                        } ;
                        if(data.configEnv.length > 0){
                            entry.env += "config:[ ";
                            data.configEnv.forEach(function(c){
                                entry.env += "{" + c.serviceName + ":" + c.type +"-"+ c.key +":"+ c.value + "},";
                            });
                            entry.env += "];   ";
                        }
                        if(data.seqEnv.length > 0){
                            entry.env += "  sequence: " + JSON.stringify(data.seqEnv) + " ;";
                        }
                        if(data.instanceEnv.length > 0){
                            entry.env += "  instance:  " + JSON.stringify(data.instanceEnv) + " ;";
                        }

                        for(var j = 0; j < result.length; j++){
                            entry.tests += result[j].className + ": " + result[j].status + "     " +getTime()+ ";   " ;
                        }
                        $scope.mixerDeltaResponse.push(entry);
                        $scope.$apply();
                    } else {
                        console.log(data.message);
                    }
                });

                stompClient.subscribe('/user/topic/mixerDeltaEnd', function (data) {
                    var data = JSON.parse(data.body);
                    console.log("\n end:");
                    console.log(data);
                    if(data.status){
                        // alert("ddminResult: " + data.ddminResult );
                        $scope.mixerDeltaResult = JSON.stringify(data.ddminResult) + "   " +getTime();
                        // console.log("data.ddminResult: " + data.ddminResult);
                        $scope.$apply();
                    } else {
                        console.log(data.message);
                    }
                    $('#test-button').removeClass('disabled');
                });

                stompClient.subscribe('/user/topic/simpleSetInstanceResult', function (data) {
                    alert(data.body);
                    $scope.refreshServices();
                    $('#test-button').removeClass('disabled');
                });


                stompClient.subscribe('/user/topic/simpleSetOrignalResult', function (data) {
                    alert(data.body);
                    $scope.refreshConfigs();
                    $('#setOriginal').removeClass('disabled');
                });

            });
        }

        function getTime(){
            var myDate = new Date();
            return myDate.getHours()+":"+myDate.getMinutes()+":"+myDate.getSeconds();
        }


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

        $scope.sendDeltaData = function(n){
            if ( stompClient != null ) {
                var checkedTest = $("input[name='testcase']:checked");
                var tests = [];
                checkedTest.each(function(){
                    tests.push($(this).val());
                });
                var checkedInstanceServices = $("input[name='instance-service']:checked");
                var instances = [];
                checkedInstanceServices.each(function(){
                    instances.push($(this).val());
                });
                var checkedSenderServices = $("input[name='sender']:checked");
                var sender = [];
                checkedSenderServices.each(function(){
                    sender.push($(this).val());
                });
                var checkedReceiverServices = $("input[name='receiver']:checked");
                var receivers = [];
                checkedReceiverServices.each(function(){
                    receivers.push($(this).val());
                });
                var checkedSenderServices2 = $("input[name='sender2']:checked");
                var sender2 = [];
                checkedSenderServices2.each(function(){
                    sender2.push($(this).val());
                });
                var checkedReceiverServices2 = $("input[name='receiver2']:checked");
                var receivers2 = [];
                checkedReceiverServices2.each(function(){
                    receivers2.push($(this).val());
                });

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

                if(tests.length > 0 && (instances.length > 0 || ((sender.length === 1 && receivers.length > 1) ||(sender2.length === 1 && receivers2.length > 1) ) || configs.length > 0)){
                    $('#test-button').addClass('disabled');
                    $scope.mixerDeltaResponse = [];
                    $scope.mixerDeltaResult = "mixer delta testing...";
                    var seqGroups = [];
                    if(sender.length == 1 && receivers.length > 1){
                        seqGroups.push({
                            'sender':sender[0],
                            'receivers': receivers
                        });
                    }
                    if(sender2.length == 1 && receivers2.length > 1){
                        seqGroups.push({
                            'sender':sender2[0],
                            'receivers': receivers2
                        });
                    }
                    var data = {
                        'id': loginId,
                        'instances': instances,
                        'seqGroups': seqGroups,
                        'configs': configs,
                        'tests':tests
                    };
                    console.log("mixers: ");
                    console.log(data);
                    stompClient.send("/app/msg/mixerDelta", {}, JSON.stringify(data));
                } else {
                    alert("Please select at least one test and one type of delta.");
                }
            } else {
                alert("Please click the connect button.")
            }
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

                if(configs.length > 0){
                    $('#setOriginal').addClass('disabled');
                    var data = {
                        'id': loginId,
                        'configs': configs
                    };
                    stompClient.send("/app/msg/simpleSetOrignal", {}, JSON.stringify(data));
                } else {
                    alert("Please select at least one config.");
                }
            } else {
                alert("Please click the connect button.")
            }
        };

        $scope.simpleSetInstance = function(n){
            if ( stompClient != null ) {
                var checkedServices = $("input[name='instance-service']:checked");
                var services = [];
                checkedServices.each(function(){
                    services.push($(this).val());
                });
                if(services.length > 0){
                    var data = {
                        'id': loginId,
                        'services': services,
                        'instanceNum': n
                    };
                    stompClient.send("/app/msg/simpleSetInstance", {}, JSON.stringify(data));
                } else {
                    alert("Please select at least one service.");
                }
            } else {
                alert("Please click the connect button.")
            }
        }



        // $scope.configlogs = "";
        // $scope.getPodLogs = function(){
        //     if ( stompClient != null ) {
        //         var checkedPods = $("input[name='pod']:checked");
        //         var pods = [];
        //         checkedPods.each(function () {
        //             pods.push($(this).val());
        //         });
        //         if(pods.length > 0){
        //             $('#inspectPodButton').addClass('disabled');
        //             getPodLogService.load(pods[0]).then(function(result){
        //                 if(result.status){
        //                     $scope.configlogs += result.podLog.podName +  ":</br>" + result.podLog.logs + "</br>";
        //                     var height = $('#config-logs').prop('scrollHeight');
        //                     $('#config-logs').scrollTop(height);
        //                 } else {
        //                     alert(result.message);
        //                 }
        //                 $('#inspectPodButton').removeClass('disabled');
        //             });
        //         } else {
        //             alert("Please check at least one pod to show its logs!");
        //         }
        //     } else {
        //         alert("Please click the connect button.")
        //     }
        //
        // };


    }]);






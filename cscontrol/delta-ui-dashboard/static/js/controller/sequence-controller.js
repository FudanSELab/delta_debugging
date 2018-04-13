var sequence = angular.module('app.sequence-controller', []);

sequence.controller('SequenceCtrl', ['$scope', '$http','$window','loadTestCases','loadServiceList', 'getPodLogService','refreshPodsService',
        function($scope, $http,$window,loadTestCases,loadServiceList, getPodLogService, refreshPodsService) {

        //刷新页面
        $scope.reloadRoute = function () {
            $window.location.reload();
        };

        // 加载service列表
        loadServiceList.loadServiceList().then(function (result) {
            if(result.status){
                $scope.services = result.services;
                $scope.senderGroup = [];
                $scope.receiverGroup = [];
                for(var i = 0; i < $scope.services.length; i++){
                    for(var j = 0; j < 5 && i < $scope.services.length; ){
                        if($scope.services[i].serviceName.indexOf("service") !== -1){
                            $scope.senderGroup.push($scope.services[i]);
                            $scope.receiverGroup.push($scope.services[i]);
                            i++;
                            j++;
                        } else {
                            i++;
                        }
                    }
                }
            } else {
                alert(result.message);
            }
        });

        // 加载testcase列表
        loadTestCases.loadTestList().then(function (result) {
            $scope.testCases = result;
        });

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

        // $scope.test = function(){
        //     var checkedTest = $("input[name='testcase']:checked");
        //     var tests = [];
        //     checkedTest.each(function(){
        //         tests.push($(this).val());
        //     });
        //     var checkedSenderServices = $("input[name='sender']:checked");
        //     var senders = [];
        //     checkedSenderServices.each(function(){
        //         senders.push($(this).val());
        //     });
        //     var checkedReceiverServices = $("input[name='receiver']:checked");
        //     var receivers = [];
        //     checkedReceiverServices.each(function(){
        //         receivers.push($(this).val());
        //     });
        //
        //     console.log(tests);
        //     console.log(senders);
        //     console.log(receivers);
        //
        // };


        var stompClient = null;
        //传递用户key值
        var loginId = new UUID().id;
        $scope.deltaResults = [];
        $scope.sequenceDeltaResult = "sequence delta testing...";

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
                stompClient.subscribe('/user/topic/sequenceDeltaResponse', function (data) {
                    var data = JSON.parse(data.body);
                    if(data.status){
                        var result = data.result.deltaResults;
                        var sender = data.sender;
                        var receivers = data.receiversInOrder;
                        var entry = {
                            sender: sender,
                            receivers:"",
                            tests: "",
                            diff:false
                        } ;
                        if(data.diffFromFirst){
                            entry.diff = true;
                        }
                        for(var i = 0; i < receivers.length; i++){
                            entry.receivers += receivers[i] + " ";
                        }
                        for(var j = 0; j < result.length; j++){
                            entry.tests += result[j].className + ": " + result[j].status + ";   " ;
                        }
                        $scope.deltaResults.push(entry);
                        $scope.$apply();
                    } else {
                        alert(data.message);
                    }

                });

                stompClient.subscribe('/user/topic/sequenceDeltaEnd', function (data) {
                    var data = JSON.parse(data.body);
                    console.log("\n end:");
                    console.log(data);
                    if(data.status){
                        alert("ddmingResult: " + data.ddminResult );
                        $scope.sequenceDeltaResult = data.ddminResult;
                        $scope.$apply();
                    } else {
                        alert(data.message);
                    }
                    $('#test-button').removeClass('disabled');
                });

            });
        }


        $scope.sendDeltaData = function() {
            var checkedTest = $("input[name='testcase']:checked");
            var tests = [];
            checkedTest.each(function(){
                tests.push($(this).val());
            });
            var checkedSenderServices = $("input[name='sender']:checked");
            var senders = [];
            checkedSenderServices.each(function(){
                senders.push($(this).val());
            });
            var checkedReceiverServices = $("input[name='receiver']:checked");
            var receivers = [];
            checkedReceiverServices.each(function(){
                receivers.push($(this).val());
            });

            if(tests.length > 0 && senders.length === 1 && receivers.length > 1){
                $('#test-button').addClass('disabled');
                $scope.deltaResults = [];
                $scope.sequenceDeltaResult = "sequence delta testing...";
                var data = {
                    'id': loginId,
                    'sender': senders[0],
                    'receivers': receivers,
                    'tests': tests
                };
                stompClient.send("/app/msg/sequenceDelta", {}, JSON.stringify(data));
            } else {
                alert("Please choose at least one testcase, one sender and two receivers.");
            }

        };

        $scope.showDelta = function(){
            $scope.deltaResults = [];
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

        $scope.instancelogs = "";
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
                            $scope.instancelogs += result.podLog.podName +  ":</br>" + result.podLog.logs + "</br>";
                            var height = $('#instance-logs').prop('scrollHeight');
                            $('#instance-logs').scrollTop(height);
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





var instance = angular.module('app.instance-controller', []);

instance.controller('InstanceCtrl', ['$scope', '$http','$window','loadTestCases','loadServiceList', 'getPodLogService','refreshPodsService',
        function($scope, $http,$window,loadTestCases,loadServiceList, getPodLogService, refreshPodsService) {

        //刷新页面
        $scope.reloadRoute = function () {
            $window.location.reload();
        };

        // 加载service列表
        loadServiceList.loadServiceList().then(function (result) {
            if(result.status){
                $scope.services = result.services;
                $scope.serviceGroup = [];
                for(var i = 0; i < $scope.services.length; i++){
                    for(var j = 0; j < 5 && i < $scope.services.length; ){
                        if($scope.services[i].serviceName.indexOf("service") !== -1){
                            // $scope.services[i].checked = false;
                            $scope.serviceGroup.push($scope.services[i]);
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
                // alert("23333");
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
                stompClient.subscribe('/user/topic/deltaresponse', function (data) {
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
                        $scope.deltaResults.push(entry);
                        $scope.$apply();
                    } else {
                        alert(data.message);
                    }

                });

                stompClient.subscribe('/user/topic/deltaend', function (data) {
                    $('#test-button').removeClass('disabled');
                    console.log( "deltaend" + data.body);
                });

            });
        }


        $scope.sendDeltaData = function() {
            $scope.deltaResults = [];
            var checkedTest = $("input[name='testcase']:checked");
            var tests = [];
            checkedTest.each(function(){
                tests.push($(this).val());
            });
            var checkedServices = $("input[name='service']:checked");
            var env = [];
            checkedServices.each(function(){
                env.push($(this).val());
            });
            console.log("tests:" + tests);
            console.log("env:" + env);

            if(tests.length > 0 && env.length > 0){
                $('#test-button').addClass('disabled');
                var data = {
                    'id': loginId,
                    'env': env,
                    'tests': tests
                };
                stompClient.send("/app/msg/delta", {}, JSON.stringify(data));
            } else {
                alert("To delete node, please select at least one service and one testcase.");
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
                        $('#inspectPodButton').removeClass('disabled');
                    } else {
                        alert(result.message);
                    }
                });
            } else {
                alert("Please check at least one pod to show its logs!");
            }
        };
        // var i = 0;
        // var timer = $interval(function () {
        //     instanceLogService.loadLogs().then(function(result){
        //         $scope.instancelogs += (++i) + ": " + result + "</br>";
        //         var height = $('#instance-logs').prop('scrollHeight');
        //         $('#instance-logs').scrollTop(height);
        //     });
        // }, 100, 30);
        //
        // timer.then(endNotify);
        //
        // function endNotify(){
        //     $scope.instancelogs += "Logs end!";
        //     var height = $('#instance-logs').prop('scrollHeight');
        //     $('#instance-logs').scrollTop(height);
        // }

}]);





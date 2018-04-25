var instance = angular.module('app.instance-controller', []);

instance.controller('InstanceCtrl', ['$scope', '$http','$window','loadTestCases','loadServiceList', 'getPodLogService','refreshPodsService','defaultCluster',
        function($scope, $http,$window,loadTestCases,loadServiceList, getPodLogService, refreshPodsService, defaultCluster) {

        //刷新页面
        $scope.reloadRoute = function () {
            $window.location.reload();
        };

        $scope.refreshServices = function(){
            loadServiceList.loadServiceList().then(function (result) {
                if(result.status){
                    $scope.services = result.services;
                    $scope.serviceGroup = [];
                    for(var i = 0; i < $scope.services.length; i++){
                        if($scope.services[i].serviceName.indexOf("service") !== -1){
                            $scope.serviceGroup.push($scope.services[i]);
                        }
                    }
                } else {
                    alert(result.message);
                }
            });
        };
        // 加载service列表
        $scope.refreshServices();


        // 加载testcase列表
        loadTestCases.loadTestList().then(function (result) {
            $scope.testCases = result;
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
        //load pods
        $scope.refreshPod();


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
                            entry.tests += result[j].className + ": " + result[j].status + "     " + new Date().toLocaleTimeString()  + ";   " ;
                        }
                        $scope.deltaResults.push(entry);
                        $scope.$apply();
                    } else {
                        alert(data.message);
                    }
                });

                stompClient.subscribe('/user/topic/deltaEnd', function (data) {
                    // console.log(data.body);
                    var data = JSON.parse(data.body);
                    if(data.status){
                        $scope.instanceDeltaResult = JSON.stringify(data.ddminResult) + "   " + new Date().toLocaleTimeString();
                        console.log("data.ddminResult: " + data.ddminResult);
                        $scope.$apply();
                    } else {
                        alert(data.message);
                    }
                    $('#test-button').removeClass('disabled');
                });

                stompClient.subscribe('/user/topic/simpleSetInstanceResult', function (data) {
                    alert(data.body);
                    $scope.refreshServices();
                    $('#test-button').removeClass('disabled');
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
                    'tests': tests,
                    'cluster': defaultCluster
                };
                stompClient.send("/app/msg/delta", {}, JSON.stringify(data));
                $scope.instanceDeltaResult = "";
            } else {
                alert("To delta instance, please select at least one service and one testcase.");
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


        $scope.simpleSetInstance = function(n){
            if ( stompClient != null ) {
                var checkedServices = $("input[name='service']:checked");
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


}]);





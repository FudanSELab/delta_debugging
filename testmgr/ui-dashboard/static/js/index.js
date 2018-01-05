var app = angular.module("myApp",[]);

app.factory('loadDataService', function ($http, $q) {

    var service = {};

    //获取并返回数据
    service.loadRecordList = function () {
        var deferred = $q.defer();
        var promise = deferred.promise;

        $http({
            method: "get",
            url: "http://localhost:5001/getFileTree",
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
        // $scope.testName = data.target.title;
        $scope.resultCount = "";
        $scope.results = [];
    });

    $scope.startTest = function(){
        if( null != $scope.testName && "" != $scope.testName){
            $http({
                method: "post",
                url: "http://localhost:5001/test",
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

    // $scope.decodeInfo = function (obj) {
    //     var des = "";
    //     for(var name in obj){
    //         des += name + ":" + obj[name] + ";";
    //     }
    //     alert(des);
    // };


});
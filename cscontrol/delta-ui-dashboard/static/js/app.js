var app = angular.module('app', ['ngRoute', 'app.instance-controller','app.node-controller', 'app.service-controller']);

app.config(['$routeProvider', function($routeProvider){
    $routeProvider
        .when('/instance', {
            templateUrl: 'templates/instance.html',
            controller: 'InstanceCtrl'
        })
        .when('/node',{
            templateUrl: 'templates/node.html',
            controller: 'NodeCtrl'
        })
        .when('/service',{
            templateUrl: 'templates/service.html',
            controller: 'ServiceCtrl'
        })
        .otherwise({redirectTo:'/instance'});
}

]);

app .controller('NavController', ['$scope', '$location', function($scope, $location) {
        $scope.isActive = function(destination) {
            // console.log( $location.path());
            return destination === $location.path();
        }
    }]);


app.directive('icheck', ['$timeout', '$parse', function($timeout, $parse) {
    return {
        restrict: 'A',
        require: '?ngModel',
        link: function(scope, element, attr, ngModel) {
            $timeout(function() {
                var value = attr.value;

                // function update(checked) {
                //     if(attr.type==='radio') {
                //         ngModel.$setViewValue(value);
                //     } else {
                //         ngModel.$setViewValue(checked);
                //     }
                // }

                $(element).iCheck({
                    checkboxClass: attr.checkboxClass || 'icheckbox_square-blue',
                    radioClass: attr.radioClass || 'iradio_square-blue'
                }).on('ifChanged', function(e) {
                    // if ($(element).attr('type') === 'checkbox' && attr['ngModel']) {
                    //     scope.$apply(function() {
                    //         return ngModel.$setViewValue(e.target.checked);
                    //     });
                    // }
                    // if ($(element).attr('type') === 'radio' && attr['ngModel']) {
                    //     return scope.$apply(function() {
                    //         return ngModel.$setViewValue(value);
                    //     });
                    // }
                });

                // scope.$watch(attr.ngChecked, function(checked) {
                //     if(typeof checked === 'undefined') checked = !!ngModel.$viewValue;
                //     update(checked)
                // }, true);
                //
                // scope.$watch(attr.ngModel, function(model) {
                //     $(element).iCheck('update');
                // }, true);

            })
        }
    }
}]);




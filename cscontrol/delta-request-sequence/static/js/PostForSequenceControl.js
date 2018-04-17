$("#set_suspend_service_btn").click(function() {
    var SetUnsetServiceRequestSuspendRequest = new Object();
    SetUnsetServiceRequestSuspendRequest.svc = $("#suspend_service_name").val();
    SetUnsetServiceRequestSuspendRequest.clusterName = "cluster5"
    var data = JSON.stringify(SetUnsetServiceRequestSuspendRequest);
    $.ajax({
        type: "post",
        url: "http://10.141.212.22:18898/api/setServiceRequestSuspend",
        contentType: "application/json",
        dataType: "json",
        data:data,
        xhrFields: {
            withCredentials: true
        },
        success: function(result){
            if(result["status"] == true){
                alert(result["message"])
            }
        },
        complete: function(){
        }
    });
});

$("#set_unsuspend_service_btn").click(function() {
    var SetUnsetServiceRequestSuspendRequest = new Object();
    SetUnsetServiceRequestSuspendRequest.svc = $("#unsuspend_service_name").val();
    SetUnsetServiceRequestSuspendRequest.clusterName = "cluster5"
    var data = JSON.stringify(SetUnsetServiceRequestSuspendRequest);
    $.ajax({
        type: "post",
        url: "http://10.141.212.22:18898/api/unsetServiceRequestSuspend",
        contentType: "application/json",
        dataType: "json",
        data:data,
        xhrFields: {
            withCredentials: true
        },
        success: function(result){
            if(result["status"] == true){
                alert(result["message"])
            }
        },
        complete: function(){
        }
    });

});

$("#service_sequence_list_suspend_all").click(function() {
    var svcListStr = $("#service_sequence_list").val();
    var svcList = svcListStr.split(",");

    for (var i = 0;i < svcList.length;i ++) {
        var SetUnsetServiceRequestSuspendRequest = new Object();
        SetUnsetServiceRequestSuspendRequest.svc = svcList[i];
        SetUnsetServiceRequestSuspendRequest.sourceSvcName = $("#source_service").val();
        SetUnsetServiceRequestSuspendRequest.clusterName = "cluster5"
        var data = JSON.stringify(SetUnsetServiceRequestSuspendRequest);
        $.ajax({
            type: "post",
            url: "http://10.141.212.22:18898/api/setServiceRequestSuspendWithSourceSvc",
            contentType: "application/json",
            dataType: "json",
            data:data,
            xhrFields: {
                withCredentials: true
            },
            success: function(result){
                if(result["status"] == true){
                    alert(result["message"])
                }
            },
            complete: function(){
            }
        });
    }
});

$("#service_sequence_list_delete_all").click(function() {
    var svcListStr = $("#service_sequence_list").val();
    var svcList = svcListStr.split(",");

    for (var i = 0;i < svcList.length;i ++) {
        var SetUnsetServiceRequestSuspendRequest = new Object();
        SetUnsetServiceRequestSuspendRequest.svc = svcList[i];
        SetUnsetServiceRequestSuspendRequest.sourceSvcName = $("#source_service").val();
        SetUnsetServiceRequestSuspendRequest.clusterName = "cluster5";
        var data = JSON.stringify(SetUnsetServiceRequestSuspendRequest);
        alert(data);
        $.ajax({
            type: "post",
            url: "http://10.141.212.22:18898/api/unsetServiceRequestSuspend",
            contentType: "application/json",
            dataType: "json",
            data:data,
            xhrFields: {
                withCredentials: true
            },
            success: function(result){
                if(result["status"] == true){
                    alert(result["message"])
                }
            },
            complete: function(){
            }
        });
    }
});


$("#service_sequence_list_check_and_unsuspend").click(function() {
    var SetAsyncRequestSequenceRequest = new Object();
    var svcList = $("#service_sequence_list").val();
    SetAsyncRequestSequenceRequest.sourceName = $("#source_service").val();
    SetAsyncRequestSequenceRequest.svcList = svcList.split(",");
    SetAsyncRequestSequenceRequest.clusterName = "cluster5";
    var data = JSON.stringify(SetAsyncRequestSequenceRequest);
    alert(data);
    $.ajax({
        type: "post",
        //url: "http://10.141.212.24:18898/api/setAsyncRequestSequenceWithSrc",
        url: "http://10.141.212.22:18898/api/setAsyncRequestSequenceWithSrcCombineWithFullSuspend",
        contentType: "application/json",
        dataType: "json",
        data:data,
        xhrFields: {
            withCredentials: true
        },
        success: function(result){
            if(result["status"] == true){
                alert(result["message"])
            }
        },
        complete: function(){
        }
    });
});
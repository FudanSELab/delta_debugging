$("#set_suspend_service_btn").click(function() {
    var SetUnsetServiceRequestSuspendRequest = new Object();
    SetUnsetServiceRequestSuspendRequest.svc = $("#suspend_service_name").val();
    SetUnsetServiceRequestSuspendRequest.actionType = 1;
    var data = JSON.stringify(SetUnsetServiceRequestSuspendRequest);
    $.ajax({
        type: "post",
        url: "/api/setServiceRequestSuspend",
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
    SetUnsetServiceRequestSuspendRequest.actionType = 2;
    var data = JSON.stringify(SetUnsetServiceRequestSuspendRequest);
    $.ajax({
        type: "post",
        url: "/api/unsetServiceRequestSuspend",
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
        SetUnsetServiceRequestSuspendRequest.svc = svcList.get(i);
        SetUnsetServiceRequestSuspendRequest.actionType = 1;
        var data = JSON.stringify(SetUnsetServiceRequestSuspendRequest);
        $.ajax({
            type: "post",
            url: "/api/setServiceRequestSuspend",
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
    SetAsyncRequestSequenceRequest.svcList = svcList.split(",");
    var data = JSON.stringify(SetAsyncRequestSequenceRequest);
    $.ajax({
        type: "post",
        url: "/api/setAsyncRequestSequence",
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
package apiserver.async;

import apiserver.request.SetAsyncRequestSequenceRequestWithSource;
import apiserver.response.SetAsyncRequestSequenceResponse;
import apiserver.service.ApiService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.scheduling.annotation.Async;
import org.springframework.scheduling.annotation.AsyncResult;
import org.springframework.stereotype.Component;
import java.util.concurrent.Future;

@Component  
public class AsyncTask {

    @Autowired
    private ApiService apiService;

    @Async("myAsync")
    public Future<SetAsyncRequestSequenceResponse> doAsync(SetAsyncRequestSequenceRequestWithSource request) throws InterruptedException{
        SetAsyncRequestSequenceResponse response =
                apiService.setAsyncRequestsSequenceWithSource(request);
        return new AsyncResult<>(response);
    }

    @Async("myAsync")
    public Future<SetAsyncRequestSequenceResponse> doAsyncWithMaintainSequence(
            SetAsyncRequestSequenceRequestWithSource request) throws InterruptedException{
        SetAsyncRequestSequenceResponse response =
                apiService.setAsyncRequestSequenceWithSrcCombineWithFullSuspendWithMaintainSequence(request);
        return new AsyncResult<>(response);
    }



}  

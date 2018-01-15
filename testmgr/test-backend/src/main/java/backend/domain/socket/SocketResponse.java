package backend.domain.socket;

import backend.domain.TestResponse;

public class SocketResponse {

    boolean status;
    String message;
    TestResponse testResult;

    public SocketResponse(){
    }

    public String getMessage() {
        return message;
    }

    public void setMessage(String message) {
        this.message = message;
    }

    public boolean getStatus() {
        return status;
    }

    public void setStatus(boolean status) {
        this.status = status;
    }

    public TestResponse getTestResult() {
        return testResult;
    }

    public void setTestResult(TestResponse testResult) {
        this.testResult = testResult;
    }

}

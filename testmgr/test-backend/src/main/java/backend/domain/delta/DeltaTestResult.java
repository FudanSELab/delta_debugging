package backend.domain.delta;

public class DeltaTestResult {

    String status; //failure/success
    String className;
    long duration = 0;//how long it run

    public DeltaTestResult(){

    }

    public String getStatus() {
        return status;
    }

    public void setStatus(String status) {
        this.status = status;
    }

    public String getClassName() {
        return className;
    }

    public void setClassName(String className) {
        this.className = className;
    }

    public long getDuration() {
        return duration;
    }

    public void setDuration(long duration) {
        this.duration = duration;
    }

}

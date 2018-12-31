package apiserver.bean;

import java.util.Map;

public class V1HealthCheck {
    private Map<String, Integer> tcpSocket;

    private int initialDelaySeconds;
    private int timeoutSeconds;
    private int periodSeconds;
    private int successThreshold;
    private int failureThreshold;

    public Map<String, Integer> getTcpSocket() {
        return tcpSocket;
    }

    public void setTcpSocket(Map<String, Integer> tcpSocket) {
        this.tcpSocket = tcpSocket;
    }

    public int getInitialDelaySeconds() {
        return initialDelaySeconds;
    }

    public void setInitialDelaySeconds(int initialDelaySeconds) {
        this.initialDelaySeconds = initialDelaySeconds;
    }

    public int getTimeoutSeconds() {
        return timeoutSeconds;
    }

    public void setTimeoutSeconds(int timeoutSeconds) {
        this.timeoutSeconds = timeoutSeconds;
    }

    public int getPeriodSeconds() {
        return periodSeconds;
    }

    public void setPeriodSeconds(int periodSeconds) {
        this.periodSeconds = periodSeconds;
    }

    public int getSuccessThreshold() {
        return successThreshold;
    }

    public void setSuccessThreshold(int successThreshold) {
        this.successThreshold = successThreshold;
    }

    public int getFailureThreshold() {
        return failureThreshold;
    }

    public void setFailureThreshold(int failureThreshold) {
        this.failureThreshold = failureThreshold;
    }
}

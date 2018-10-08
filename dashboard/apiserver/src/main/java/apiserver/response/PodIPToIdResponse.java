package apiserver.response;

import java.util.Map;

public class PodIPToIdResponse {
    private Boolean status;
    private String message;
    private Map<String, String> ipToIdMap; // key is ip, value is id

    public Boolean getStatus() {
        return status;
    }

    public void setStatus(Boolean status) {
        this.status = status;
    }

    public String getMessage() {
        return message;
    }

    public void setMessage(String message) {
        this.message = message;
    }

    public Map<String, String> getIpToIdMap() {
        return ipToIdMap;
    }

    public void setIpToIdMap(Map<String, String> ipToIdMap) {
        this.ipToIdMap = ipToIdMap;
    }
}

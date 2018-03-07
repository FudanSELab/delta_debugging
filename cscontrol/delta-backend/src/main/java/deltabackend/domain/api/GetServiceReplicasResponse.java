package deltabackend.domain.api;

import deltabackend.domain.EnvParameter;

import java.util.List;

public class GetServiceReplicasResponse {
    private boolean status;
    private String message;
    private List<EnvParameter> services;

    public GetServiceReplicasResponse(){

    }

    public boolean isStatus() {
        return status;
    }

    public void setStatus(boolean status) {
        this.status = status;
    }

    public String getMessage() {
        return message;
    }

    public void setMessage(String message) {
        this.message = message;
    }

    public List<EnvParameter> getServices() {
        return services;
    }

    public void setServices(List<EnvParameter> services) {
        this.services = services;
    }

    @Override
    public String toString(){
        String s = "";
        for(EnvParameter ep : services){
            s += ep.getServiceName() + ": " + ep.getNumOfReplicas() + "  ";
        }
        return s;
    }


}

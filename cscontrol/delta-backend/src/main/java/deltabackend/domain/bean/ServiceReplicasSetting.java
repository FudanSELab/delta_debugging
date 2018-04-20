package deltabackend.domain.bean;

public class ServiceReplicasSetting implements Cloneable{
    private String serviceName;
    private int numOfReplicas;

    public ServiceReplicasSetting(){

    }

    public String getServiceName() {
        return serviceName;
    }

    public void setServiceName(String serviceName) {
        this.serviceName = serviceName;
    }

    public int getNumOfReplicas() {
        return numOfReplicas;
    }

    public void setNumOfReplicas(int numOfReplicas) {
        this.numOfReplicas = numOfReplicas;
    }

    @Override
    public Object clone() {
        ServiceReplicasSetting p = null;
        try{
            p = (ServiceReplicasSetting)super.clone();
        }catch(CloneNotSupportedException e) {
            e.printStackTrace();
        }
        p.setServiceName(this.serviceName);
        p.setNumOfReplicas(this.numOfReplicas);
        return p;
    }
}

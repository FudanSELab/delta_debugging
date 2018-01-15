package deltabackend.domain;

public class EnvParameter implements Cloneable{

    String serviceName;
    int instanceNum;

    public String getServiceName() {
        return serviceName;
    }

    public void setServiceName(String serviceName) {
        this.serviceName = serviceName;
    }

    public int getInstanceNum() {
        return instanceNum;
    }

    public void setInstanceNum(int instanceNum) {
        this.instanceNum = instanceNum;
    }

    @Override
    public Object clone() {
        EnvParameter p = null;
        try{
            p = (EnvParameter)super.clone();
        }catch(CloneNotSupportedException e) {
            e.printStackTrace();
        }
        p.serviceName = this.serviceName;
        p.instanceNum = this.instanceNum;
        return p;
    }

}

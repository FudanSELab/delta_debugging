package apiserver.bean;

import java.util.Objects;

public class SetServicesReplicasResponseFromAPI {
    private String apiVersion = null;

    private String kind = null;

    private V1ObjectMeta metadata = null;

    private AppsV1beta1DeploymentSpec spec = null;

    private AppsV1beta1DeploymentStatus status = null;

    public SetServicesReplicasResponseFromAPI apiVersion(String apiVersion) {
        this.apiVersion = apiVersion;
        return this;
    }

    public String getApiVersion() {
        return apiVersion;
    }

    public void setApiVersion(String apiVersion) {
        this.apiVersion = apiVersion;
    }

    public SetServicesReplicasResponseFromAPI kind(String kind) {
        this.kind = kind;
        return this;
    }


    public String getKind() {
        return kind;
    }

    public void setKind(String kind) {
        this.kind = kind;
    }

    public SetServicesReplicasResponseFromAPI metadata(V1ObjectMeta metadata) {
        this.metadata = metadata;
        return this;
    }

    public V1ObjectMeta getMetadata() {
        return metadata;
    }

    public void setMetadata(V1ObjectMeta metadata) {
        this.metadata = metadata;
    }

    public SetServicesReplicasResponseFromAPI spec(AppsV1beta1DeploymentSpec spec) {
        this.spec = spec;
        return this;
    }

    public AppsV1beta1DeploymentSpec getSpec() {
        return spec;
    }

    public void setSpec(AppsV1beta1DeploymentSpec spec) {
        this.spec = spec;
    }

    public SetServicesReplicasResponseFromAPI status(AppsV1beta1DeploymentStatus status) {
        this.status = status;
        return this;
    }

    public AppsV1beta1DeploymentStatus getStatus() {
        return status;
    }

    public void setStatus(AppsV1beta1DeploymentStatus status) {
        this.status = status;
    }


    @Override
    public boolean equals(java.lang.Object o) {
        if (this == o) {
            return true;
        }
        if (o == null || getClass() != o.getClass()) {
            return false;
        }
        SetServicesReplicasResponseFromAPI appsV1beta1Deployment = (SetServicesReplicasResponseFromAPI) o;
        return Objects.equals(this.apiVersion, appsV1beta1Deployment.apiVersion) &&
                Objects.equals(this.kind, appsV1beta1Deployment.kind) &&
                Objects.equals(this.metadata, appsV1beta1Deployment.metadata) &&
                Objects.equals(this.spec, appsV1beta1Deployment.spec) &&
                Objects.equals(this.status, appsV1beta1Deployment.status);
    }

    @Override
    public int hashCode() {
        return Objects.hash(apiVersion, kind, metadata, spec, status);
    }


    @Override
    public String toString() {
        StringBuilder sb = new StringBuilder();
        sb.append("class AppsV1beta1Deployment {\n");

        sb.append("    apiVersion: ").append(toIndentedString(apiVersion)).append("\n");
        sb.append("    kind: ").append(toIndentedString(kind)).append("\n");
        sb.append("    metadata: ").append(toIndentedString(metadata)).append("\n");
        sb.append("    spec: ").append(toIndentedString(spec)).append("\n");
        sb.append("    status: ").append(toIndentedString(status)).append("\n");
        sb.append("}");
        return sb.toString();
    }

    /**
     * Convert the given object to string with each line indented by 4 spaces
     * (except the first line).
     */
    private String toIndentedString(java.lang.Object o) {
        if (o == null) {
            return "null";
        }
        return o.toString().replace("\n", "\n    ");
    }

}

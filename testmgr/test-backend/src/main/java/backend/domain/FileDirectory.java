package backend.domain;

import java.util.ArrayList;

public class FileDirectory extends FileNode {

    ArrayList<FileNode> products = new ArrayList<FileNode>();

    public FileDirectory(){

    }

    public ArrayList<FileNode> getProducts() {
        return products;
    }

    public void setProducts(ArrayList<FileNode> products) {
        this.products = products;
    }

    public void addProduct(FileNode fn){
        products.add(fn);
    }

}

package backend.domain;

import java.io.*;

public class MyClassLoader extends ClassLoader{
    private String rootUrl;//http://localhost:8090/Test.class

    public MyClassLoader(String rootUrl) {
        super();
        this.rootUrl = rootUrl;
    }

    @Override
    protected Class<?> findClass(String name) throws ClassNotFoundException {
        Class clazz = null;
        //this.findLoadedClass(name); // 父类已加载
        //if (clazz == null) {  //检查该类是否已被加载过
        byte[] classData = getClassData(name);  //根据类的二进制名称,获得该class文件的字节码数组
        if (classData == null) {
            throw new ClassNotFoundException();
        }
        clazz = defineClass(name, classData, 0, classData.length);  //将class的字节码数组转换成Class类的实例
        //}
        return clazz;
    }

    private byte[] getClassData(String name) {
        InputStream is = null;
        try {
            String path = classNameToPath(name);
            File f = new File(path);
            byte[] buff = new byte[1024*4];
            int len = -1;
            is = new FileInputStream(f);
            ByteArrayOutputStream baos = new ByteArrayOutputStream();
            while((len = is.read(buff)) != -1) {
                baos.write(buff,0,len);
            }
            return baos.toByteArray();
        } catch (Exception e) {
            e.printStackTrace();
        } finally {
            if (is != null) {
                try {
                    is.close();
                } catch(IOException e) {
                    e.printStackTrace();
                }
            }
        }
        return null;
    }

    private String classNameToPath(String name) {
        System.out.println(rootUrl + "/" + name.replace(".", "/") + ".class");
        return rootUrl + "/" + name.replace(".", "/") + ".class";
    }
}

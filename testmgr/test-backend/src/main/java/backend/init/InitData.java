package backend.init;

import backend.service.ConfigService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.CommandLineRunner;

public class InitData implements CommandLineRunner {

    @Autowired
    ConfigService configService;

    @Override
    public void run(String... strings) throws Exception {
        configService.init();
    }
}

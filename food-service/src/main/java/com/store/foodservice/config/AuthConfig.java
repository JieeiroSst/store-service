package com.store.foodservice.config;

import com.store.foodservice.auth.AuthFilter;
import com.store.foodservice.auth.Authenticator;
import org.springframework.boot.web.servlet.FilterRegistrationBean;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class AuthConfig {
    @Bean
    public FilterRegistrationBean<AuthFilter> authFilter() {
        FilterRegistrationBean<AuthFilter> bean = new FilterRegistrationBean<>(new AuthFilter(Authenticator.fromEnv()));
        bean.addUrlPatterns("/api/*");
        bean.setOrder(1);
        return bean;
    }
}

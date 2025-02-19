package com.store.foodservice.config;

import java.time.Duration;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.data.redis.cache.RedisCacheConfiguration;
import org.springframework.data.redis.cache.RedisCacheManager;
//import org.springframework.data.redis.connection.RedisConnectionFactory;
import org.springframework.data.redis.connection.RedisStandaloneConfiguration;
import org.springframework.data.redis.connection.lettuce.LettuceConnectionFactory;
import org.springframework.data.redis.serializer.GenericJackson2JsonRedisSerializer;
import org.springframework.data.redis.serializer.RedisSerializationContext.SerializationPair;

@Configuration
public class RedisConfig {
    @Value("${redis.host}")
    private String redisHost;

    @Value("${redis.port}")
    private int redisPort;

    @Value("${redis.user}")
    private String redisUser;

    @Value("${redis.password}")
    private String redisPassword;

    @Bean
    public LettuceConnectionFactory redisConnectionFactory() {
        RedisStandaloneConfiguration configuration = new RedisStandaloneConfiguration(redisHost, redisPort);
        configuration.setUsername(redisUser);
        configuration.setPassword(redisPassword);

        return new LettuceConnectionFactory(configuration);
    }

    // @Bean
    // public RedisCacheManager cacheManager(RedisConnectionFactory
    // connectionFactory) {
    // return RedisCacheManager.create(connectionFactory);
    // }

    @Bean
    public RedisCacheManager cacheManager() {
        RedisCacheConfiguration cacheConfig = myDefaultCacheConfig(Duration.ofMinutes(10)).disableCachingNullValues();

        return RedisCacheManager.builder(redisConnectionFactory())
                .cacheDefaults(cacheConfig)
                .withCacheConfiguration("ingredients", myDefaultCacheConfig(Duration.ofMinutes(5)))
                .withCacheConfiguration("ingredient", myDefaultCacheConfig(Duration.ofMinutes(1)))
                .withCacheConfiguration("food", myDefaultCacheConfig(Duration.ofMinutes(5)))
                .withCacheConfiguration("foods", myDefaultCacheConfig(Duration.ofMinutes(1)))
                .withCacheConfiguration("allergen", myDefaultCacheConfig(Duration.ofMinutes(5)))
                .withCacheConfiguration("allergens", myDefaultCacheConfig(Duration.ofMinutes(1)))
                .build();
    }

    private RedisCacheConfiguration myDefaultCacheConfig(Duration duration) {
        return RedisCacheConfiguration
                .defaultCacheConfig()
                .entryTtl(duration)
                .serializeValuesWith(SerializationPair.fromSerializer(new GenericJackson2JsonRedisSerializer()));
    }
}

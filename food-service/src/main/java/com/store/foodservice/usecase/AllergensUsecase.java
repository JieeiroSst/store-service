package com.store.foodservice.usecase;

import java.util.List;
import java.util.Optional;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.cache.annotation.CacheEvict;
import org.springframework.cache.annotation.Cacheable;
import org.springframework.cache.annotation.EnableCaching;
import org.springframework.stereotype.Service;

import com.store.foodservice.model.Allergens;
import com.store.foodservice.repository.AllergensRepository;

@Service
@EnableCaching
public class AllergensUsecase {
    @Autowired
    AllergensRepository allergensRepository;

    private void doLongRunningTask() {
        try {
            Thread.sleep(3000);
        } catch (InterruptedException e) {
            e.printStackTrace();
        }
    }
}

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

    @Cacheable("allergens")
    public List<Allergens> findAll() {
        doLongRunningTask();

        return allergensRepository.findAll();
    }

    public Allergens save(Allergens req) {
        return allergensRepository.save(req);
    }

    @CacheEvict(value = "allergen", key = "#allergen.id")
    public Allergens update(Allergens req) {
        return allergensRepository.save(req);
    }

    @CacheEvict(value = "allergen", key = "#id")
    public void deleteById(long id) {
        allergensRepository.deleteById(id);
    }

    @Cacheable("allergen")
    public Optional<Allergens> findById(long id) {
        doLongRunningTask();

        return allergensRepository.findById(id);
    }

    @CacheEvict(value = { "allergen", "allergens" }, allEntries = true)
    public void deleteAll() {
        allergensRepository.deleteAll();
    }

    private void doLongRunningTask() {
        try {
            Thread.sleep(3000);
        } catch (InterruptedException e) {
            e.printStackTrace();
        }
    }
}

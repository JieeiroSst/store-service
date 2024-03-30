package com.store.foodservice.usecase;

import java.util.List;
import java.util.Optional;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.cache.annotation.CacheEvict;
import org.springframework.cache.annotation.Cacheable;
import org.springframework.cache.annotation.EnableCaching;
import org.springframework.stereotype.Service;

import com.store.foodservice.repository.FoodRepository;
import com.store.foodservice.model.Food;

@Service
@EnableCaching
public class FoodUsecase {
    @Autowired
    FoodRepository foodRepository;

    @Cacheable("foods")
    public List<Food> findAll() {
        doLongRunningTask();

        return foodRepository.findAll();
    }

    public Food save(Food req) {
        return foodRepository.save(req);
    }

    @CacheEvict(value = "food", key = "#food.id")
    public Food update(Food req) {
        return foodRepository.save(req);
    }

    @CacheEvict(value = "food", key = "#id")
    public void deleteById(long id) {
        foodRepository.deleteById(id);
    }

    @Cacheable("food")
    public Optional<Food> findById(long id) {
        doLongRunningTask();

        return foodRepository.findById(id);
    }

    @CacheEvict(value = { "food", "foods" }, allEntries = true)
    public void deleteAll() {
        foodRepository.deleteAll();
    }

    @Cacheable("name_foods")
    public List<Food> findByNameContaining(String name) {
        doLongRunningTask();

        return foodRepository.findByNameContaining(name);
    }

    private void doLongRunningTask() {
        try {
            Thread.sleep(3000);
        } catch (InterruptedException e) {
            e.printStackTrace();
        }
    }
}

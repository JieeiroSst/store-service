package com.store.foodservice.usecase;

import java.util.List;
import java.util.Optional;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.cache.annotation.CacheEvict;
import org.springframework.cache.annotation.Cacheable;
import org.springframework.cache.annotation.EnableCaching;
import org.springframework.stereotype.Service;

import com.store.foodservice.model.Ingredient;
import com.store.foodservice.repository.IngredientRepository;

@Service
@EnableCaching
public class IngredientUsecase {
    @Autowired
    IngredientRepository ingredientRepository;

    @Cacheable("ingredients")
    public List<Ingredient> findAll() {
        doLongRunningTask();

        return ingredientRepository.findAll();
    }

    public Ingredient save(Ingredient req) {
        return ingredientRepository.save(req);
    }

    @CacheEvict(value = "ingredient", key = "#ingredient.id")
    public Ingredient update(Ingredient req) {
        return ingredientRepository.save(req);
    }

    @CacheEvict(value = "ingredient", key = "#id")
    public void deleteById(long id) {
        ingredientRepository.deleteById(id);
    }

    @Cacheable("ingredient")
    public Optional<Ingredient> findById(long id) {
        doLongRunningTask();

        return ingredientRepository.findById(id);
    }

    @CacheEvict(value = { "ingredient", "ingredients" }, allEntries = true)
    public void deleteAll() {
        ingredientRepository.deleteAll();
    }

    private void doLongRunningTask() {
        try {
            Thread.sleep(3000);
        } catch (InterruptedException e) {
            e.printStackTrace();
        }
    }
}

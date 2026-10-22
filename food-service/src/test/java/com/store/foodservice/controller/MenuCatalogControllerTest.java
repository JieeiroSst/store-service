package com.store.foodservice.controller;

import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

import java.util.List;
import java.util.Optional;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;

import com.store.foodservice.model.Food;
import com.store.foodservice.model.Restaurant;
import com.store.foodservice.repository.FoodRepository;
import com.store.foodservice.repository.RestaurantRepository;

class MenuCatalogControllerTest {
    private RestaurantRepository restaurants;
    private FoodRepository foods;
    private MockMvc mvc;

    @BeforeEach
    void setUp() {
        restaurants = mock(RestaurantRepository.class);
        foods = mock(FoodRepository.class);
        mvc = MockMvcBuilders.standaloneSetup(
                new MenuCatalogController(restaurants, foods),
                new RestaurantController(restaurants, foods)).build();
    }

    private static Restaurant restaurant(long id, boolean active) {
        Restaurant r = new Restaurant("Phở Thìn", "13 Lò Đúc");
        r.setId(id);
        r.setActive(active);
        r.setCommissionRate(0.15);
        return r;
    }

    private static Food food(long id, Long restaurantId, Boolean active) {
        Food f = new Food("Phở bò", "Tái chín", 55000, "Món nước");
        f.setId(id);
        f.setRestaurantId(restaurantId);
        f.setActive(active);
        return f;
    }

    @Test
    void restaurantMatchesDoordashContract() throws Exception {
        when(restaurants.findById(5L)).thenReturn(Optional.of(restaurant(5, true)));
        mvc.perform(get("/v1/restaurants/5"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.restaurant_id").value("5"))
                .andExpect(jsonPath("$.name").value("Phở Thìn"))
                .andExpect(jsonPath("$.is_active").value(true))
                .andExpect(jsonPath("$.commission_rate").value(0.15));
    }

    @Test
    void unknownOrMalformedIdsAreNotFound() throws Exception {
        when(restaurants.findById(9L)).thenReturn(Optional.empty());
        mvc.perform(get("/v1/restaurants/9")).andExpect(status().isNotFound());
        mvc.perform(get("/v1/restaurants/keiki-kitchen")).andExpect(status().isNotFound());
        mvc.perform(get("/v1/menu-items/-1")).andExpect(status().isNotFound());
    }

    @Test
    void menuItemMatchesDoordashContract() throws Exception {
        when(foods.findById(11L)).thenReturn(Optional.of(food(11, 5L, null)));
        mvc.perform(get("/v1/menu-items/11"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.item_id").value("11"))
                .andExpect(jsonPath("$.restaurant_id").value("5"))
                .andExpect(jsonPath("$.name").value("Phở bò"))
                .andExpect(jsonPath("$.price").value(55000.0))
                .andExpect(jsonPath("$.is_active").value(true));
    }

    @Test
    void foodsWithoutARestaurantCannotBeOrdered() throws Exception {
        when(foods.findById(12L)).thenReturn(Optional.of(food(12, null, true)));
        mvc.perform(get("/v1/menu-items/12"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.is_active").value(false));
    }

    @Test
    void restaurantListAndMenuHideInactiveEntries() throws Exception {
        when(restaurants.findAll()).thenReturn(List.of(restaurant(1, true), restaurant(2, false)));
        when(restaurants.existsById(1L)).thenReturn(true);
        when(foods.findByRestaurantId(1L)).thenReturn(List.of(food(3, 1L, true), food(4, 1L, false)));

        mvc.perform(get("/api/restaurants"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.length()").value(1))
                .andExpect(jsonPath("$[0].id").value(1))
                .andExpect(jsonPath("$[0].is_active").value(true));
        mvc.perform(get("/api/restaurants/1/foods"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.length()").value(1))
                .andExpect(jsonPath("$[0].id").value(3))
                .andExpect(jsonPath("$[0].restaurant_id").value(1));
    }

    @Test
    void createRequiresAName() throws Exception {
        mvc.perform(post("/api/restaurants").contentType(MediaType.APPLICATION_JSON).content("{\"address\":\"x\"}"))
                .andExpect(status().isBadRequest());
    }
}

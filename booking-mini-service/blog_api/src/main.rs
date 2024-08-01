use blog_shared::Post;
fn main() {
    let post = Post::new(
        "1234".to_owned(),
    );

    println!("{post:?}");
}
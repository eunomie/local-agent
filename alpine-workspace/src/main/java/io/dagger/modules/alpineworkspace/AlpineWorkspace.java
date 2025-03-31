package io.dagger.modules.alpineworkspace;

import static io.dagger.client.Dagger.dag;

import io.dagger.client.*;
import io.dagger.module.annotation.Function;
import io.dagger.module.annotation.Object;
import java.util.List;
import java.util.concurrent.ExecutionException;

@Object
public class AlpineWorkspace {
  public Container container;

  public AlpineWorkspace() {}

  public AlpineWorkspace(Directory source) {
    this.container =
        dag()
            .container()
            .from("alpine:3")
            .withDirectory("/workspace", source)
            .withWorkdir("/workspace");
  }

  /**
   * Install system packages using apk.
   *
   * This is not suitable to install project dependencies.
   *
   * @param packages List of alpine packages to install
   */
  @Function
  public AlpineWorkspace addPackages(List<String> packages)
      throws ExecutionException, DaggerQueryException, InterruptedException {
    List<String> args = new java.util.ArrayList<>();
    args.addAll(List.of("apk", "add", "--no-cache"));
    args.addAll(packages);
    return withExec(args);
  }

  /**
   * Run shell command.
   *
   * This is useful to install project dependencies, run tests, etc.
   *
   * @param args Command to run
   */
  @Function
  public AlpineWorkspace withExec(List<String> args)
      throws ExecutionException, DaggerQueryException, InterruptedException {
    this.container =
        this.container.withExec(args, new Container.WithExecArguments().withExpect(ReturnType.ANY));
    String out = this.container.stdout() + "\n" + this.container.stderr();
    int exitCode = this.container.exitCode();
    if (exitCode != 0) {
      throw new RuntimeException("Failed to execute command: " + out);
    }
    return this;
  }

  /**
   * Read the contents of a file at the given path
   *
   * @param path Path to read the file at
   */
  @Function
  public String read(String path)
      throws ExecutionException, DaggerQueryException, InterruptedException {
    return container.file(path).contents();
  }

  /**
   * Write the contents of a file at the given path
   *
   * @param path Path to write the file at
   * @param contents Contents to write
   */
  @Function
  public AlpineWorkspace write(String path, String contents) {
    this.container = this.container.withNewFile(path, contents);
    return this;
  }

  /** List the available files in tree format */
  @Function
  public String tree() throws ExecutionException, DaggerQueryException, InterruptedException {
    return this.container.withExec(List.of("tree", ".")).stdout();
  }
}

## Refactoring

### Clean Code

- The main purpose of `Refactoring` is to fight technical debt. It transform a mess into clean code and simple design.

  - **1. Clean code is obvious for other programming**: Poor variable naming, bloated classes and methods, magic numbers, all of that makes code `sloppy` and `difficult to grasp`.

  - **2. Clean code doesn't contain duplication**

  - **3. Clean code contains a minimal number of classes and other moving parts**: Less code is less stuff to keep in your head. Less code is fewer bugs. Code is liability, keep it short and simple

  - **4. Clean code passes all tests**

  - **5. Clean code is easier and cheaper to maintain**

### Technical Debt

- Causes of technical debt:
  
  - **1. Business pressure**: Sometime business circumstances might force you to roll out features before they're completely finished. In this case, patches and kludges will appear in the code to hide the unfinished parts of the project.

  - **2. Lack of understanding of the consequences of technical debt**: Sometimes your employer might not understand that technical debt has “interest” insofar as it slows down the pace of development as debt accumulates. This can make it too difficult to dedicate the team’s time to refactoring because management doesn't see the value of it.

  - **3. Failing to combat the strict coherence of components**: This is when the project resembles a monolith rather than the product of individual modules. In this case, any changes to one part of the project will affect others. Team development is made more difficult because it’s difficult to isolate the work of individual members.

  - **4. Lack of tests**: The lack of immediate feedback encourages quick, but risky workarounds or kludges. In worst cases, these changes are implemented and deployed right into the production without any prior testing. The consequences can be catastrophic. For example, an innocent-looking hot fix might send a weird test email to thousands of customers or even worse, flush or corrupt an entire database.

  - **5. Lack of documentation**: This slows down the introduction of new people to the project and can grind development to a halt if key people leave the project.

  - **6. Lack of interaction between team members**: If the knowledge base isn't distributed throughout the company, people will end up working with an outdated understanding of processes and information about the project. This situation can be exacerbated when junior developers are incorrectly trained by their mentors.

  - **7. Long-term simultaneous development in several branches**: This can lead to the accumulation of technical debt, which is then increased when changes are merged. The more changes made in isolation, the greater the total technical debt.

  - **8. Delayed refactoring**: The project’s requirements are constantly changing and at some point it may become obvious that parts of the code are obsolete, have become cumbersome, and must be redesigned to meet new requirements. On the other hand, the project’s programmers are writing new code every day that works with the obsolete parts. Therefore, the longer refactoring is delayed, the more dependent code will have to be reworked in the future.

  - **9. Lack of compliance monitoring**: This happens when everyone working on the project writes code as they see fit (i.e. the same way they wrote the last project).

  - **10. Incompetence**: This is when the developer just doesn't know how to write decent code.
